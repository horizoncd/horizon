package cluster

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"text/template"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	clustercommon "g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/registry"
	clustertagmanager "g.hz.netease.com/horizon/pkg/clustertag/manager"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/jsonschema"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"github.com/Masterminds/sprig"
	"github.com/go-yaml/yaml"
)

func (c *controller) ListCluster(ctx context.Context, applicationID uint, environments []string,
	filter string, query *q.Query) (_ int, _ []*ListClusterResponse, err error) {
	const op = "cluster controller: list cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	count, clustersWithEnvAndRegion, err := c.clusterMgr.ListByApplicationAndEnvs(ctx,
		applicationID, environments, filter, query)
	if err != nil {
		return 0, nil, errors.E(op, err)
	}

	return count, ofClustersWithEnvAndRegion(clustersWithEnvAndRegion), nil
}

func (c *controller) ListClusterByNameFuzzily(ctx context.Context, environment,
	filter string, query *q.Query) (count int, listClusterWithFullResp []*ListClusterWithFullResponse, err error) {
	const op = "cluster controller: list cluster by name fuzzily"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	listClusterWithFullResp = []*ListClusterWithFullResponse{}
	// 1. get clusters
	count, clustersWithEnvAndRegion, err := c.clusterMgr.ListByNameFuzzily(ctx,
		environment, filter, query)
	if err != nil {
		return 0, nil, errors.E(op, err)
	}

	// 2. get applications
	var applicationIDs []uint
	for _, cluster := range clustersWithEnvAndRegion {
		applicationIDs = append(applicationIDs, cluster.ApplicationID)
	}
	applicationMap, err := c.applicationSvc.GetByIDs(ctx, applicationIDs)
	if err != nil {
		return 0, nil, errors.E(op, err)
	}

	// 3. convert and add full path, full name
	for _, cluster := range clustersWithEnvAndRegion {
		application, exist := applicationMap[cluster.ApplicationID]
		if !exist {
			continue
		}
		fullPath := fmt.Sprintf("%v/%v", application.FullPath, cluster.Name)
		fullName := fmt.Sprintf("%v/%v", application.FullName, cluster.Name)
		listClusterWithFullResp = append(listClusterWithFullResp, &ListClusterWithFullResponse{
			ofClusterWithEnvAndRegion(cluster),
			fullName,
			fullPath,
		})
	}

	return count, listClusterWithFullResp, nil
}

func (c *controller) GetCluster(ctx context.Context, clusterID uint) (_ *GetClusterResponse, err error) {
	const op = "cluster controller: get cluster"
	l := wlog.Start(ctx, op)
	defer func() {
		// errors like ClusterNotFound are logged with info level
		if err != nil && errors.Status(err) == http.StatusNotFound {
			log.WithFiled(ctx, "op",
				op).WithField("duration", l.GetDuration().String()).Info(wlog.ByErr(err))
		} else {
			l.Stop(func() string { return wlog.ByErr(err) })
		}
	}()

	// 1. get cluster from db
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. get application
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. get environmentRegion
	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 4. get files in git repo
	clusterFiles, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 5. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	// 6. transfer model
	clusterResp := ofClusterModel(application, cluster, er, fullPath,
		clusterFiles.PipelineJSONBlob, clusterFiles.ApplicationJSONBlob)

	// 7. get latest deployed commit
	latestPR, err := c.pipelinerunMgr.GetLatestSuccessByClusterID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if latestPR != nil {
		clusterResp.LatestDeployedCommit = latestPR.GitCommit
	}

	// 8. get createdBy and updatedBy users
	userMap, err := c.userManager.GetUserMapByIDs(ctx, []uint{cluster.CreatedBy, cluster.UpdatedBy})
	if err != nil {
		return nil, errors.E(op, err)
	}
	clusterResp.CreatedBy = toUser(getUserFromMap(cluster.CreatedBy, userMap))
	clusterResp.UpdatedBy = toUser(getUserFromMap(cluster.UpdatedBy, userMap))

	return clusterResp, nil
}

func (c *controller) GetClusterOutput(ctx context.Context, clusterID uint) (_ string, err error) {
	const op = "cluster controller: get cluster output"
	defer wlog.Start(ctx, op).StopPrint()
	// 1. get cluster from db
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		if errors.Status(err) != http.StatusNotFound {
			log.Errorf(ctx, "get cluster error, err = %s", err.Error())
		}
		return "", errors.E(op, err)
	}

	// 2. get application
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		if errors.Status(err) != http.StatusNotFound {
			log.Errorf(ctx, "get application error, err = %s", err.Error())
		}
		return "", errors.E(op, err)
	}

	// 3. get output in template
	outputStr, err := c.outputGetter.GetTemplateOutPut(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		log.Errorf(ctx, "get template output error, err = %s", err.Error())
		return "", err
	}
	if outputStr == "" {
		return "", nil
	}

	// 4. get files in  git repo
	clusterFiles, err := c.clusterGitRepo.GetClusterValueFiles(ctx, application.Name, cluster.Name)
	if err != nil {
		log.Errorf(ctx, "get clusterValueFile from gitRepo error, err  = %s", err.Error())
		return "", err
	}

	log.Debugf(ctx, "clusterFiles = %+v, outputStr = %+v", clusterFiles, outputStr)

	// 5. reader output in template and return
	outputRenderStr, err := RenderOutPutStr(outputStr, cluster.Template, clusterFiles...)
	if err != nil {
		log.Errorf(ctx, "render outputstr error, err = %s", err.Error())
		return "", err
	}

	return outputRenderStr, nil
}

func RenderOutPutStr(outPutStr, templateName string, clusterValueFiles ...gitrepo.ClusterValueFile) (string, error) {
	// remove the  template prefix level, and merge to on doc
	var oneDoc string
	for _, clusterValueFile := range clusterValueFiles {
		if clusterValueFile.Content != nil {
			if content, ok := clusterValueFile.Content[templateName]; ok {
				binaryContent, err := yaml.Marshal(content)
				if err != nil {
					return "", err
				}
				oneDoc += string(binaryContent) + "\n"
			}
		}
	}
	var oneDocMap map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(oneDoc), &oneDocMap)
	if err != nil {
		return "", err
	}

	// template the outPutStr
	var b bytes.Buffer
	doTemplate := template.Must(template.New("").Funcs(sprig.TxtFuncMap()).Parse(outPutStr))
	err = doTemplate.ExecuteTemplate(&b, "", oneDocMap)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

func (c *controller) CreateCluster(ctx context.Context, applicationID uint,
	environment, region string, extraOwners []string, r *CreateClusterRequest) (_ *GetClusterResponse, err error) {
	const op = "cluster controller: create cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. get application
	application, err := c.applicationMgr.GetByID(ctx, applicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. validate
	exists, err := c.clusterMgr.CheckClusterExists(ctx, r.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if exists {
		return nil, errors.E(op, http.StatusConflict, errors.ErrorCode("Conflict"), "已存在同名集群，请勿重复创建！")
	}
	if err := c.validateCreate(r); err != nil {
		return nil, errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}

	err = c.userSvc.CheckUsersExists(ctx, extraOwners)
	if err != nil {
		return nil, errors.E(op, http.StatusBadRequest, errors.ErrorCode(common.InvalidRequestParam), err)
	}

	// 3. if template is empty, set it with application's template
	if r.Template == nil {
		r.Template = &Template{
			Name:    application.Template,
			Release: application.TemplateRelease,
		}
	} else {
		if r.Template.Name == "" {
			r.Template.Name = application.Template
		}
		if r.Template.Release == "" {
			r.Template.Release = application.TemplateRelease
		}
	}

	// 4. if templateInput is empty, set it with application's templateInput
	if r.TemplateInput == nil {
		pipelineJSONBlob, applicationJSONBlob, err := c.applicationGitRepo.GetApplication(ctx, application.Name)
		if err != nil {
			return nil, errors.E(op, err)
		}
		r.TemplateInput = &TemplateInput{}
		r.TemplateInput.Application = applicationJSONBlob
		r.TemplateInput.Pipeline = pipelineJSONBlob
	} else {
		if err := c.validateTemplateInput(ctx, r.Template.Name,
			r.Template.Release, r.TemplateInput, nil); err != nil {
			return nil, errors.E(op, http.StatusBadRequest,
				errors.ErrorCode(common.InvalidRequestBody), err)
		}
	}

	// 5. get environmentRegion
	er, err := c.envMgr.GetByEnvironmentAndRegion(ctx, environment, region)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 6. get regionEntity
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, region)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 7. get templateRelease
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, r.Template.Name, r.Template.Release)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 8. create cluster, after created, params.Cluster is the newest cluster
	cluster, clusterTags := r.toClusterModel(application, er)
	cluster.CreatedBy = currentUser.GetID()
	cluster.UpdatedBy = currentUser.GetID()

	if err := clustertagmanager.ValidateUpsert(clusterTags); err != nil {
		return nil, errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}

	// 9. create cluster in git repo
	err = c.clusterGitRepo.CreateCluster(ctx, &gitrepo.CreateClusterParams{
		BaseParams: &gitrepo.BaseParams{
			Cluster:             cluster.Name,
			PipelineJSONBlob:    r.TemplateInput.Pipeline,
			ApplicationJSONBlob: r.TemplateInput.Application,
			TemplateRelease:     tr,
			Application:         application,
			Environment:         environment,
		},
		RegionEntity: regionEntity,
		ClusterTags:  clusterTags,
		Namespace:    r.Namespace,
		Image:        r.Image,
	})
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 10. create cluster in db
	cluster, err = c.clusterMgr.Create(ctx, cluster, clusterTags, extraOwners)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 11. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	ret := ofClusterModel(application, cluster, er, fullPath,
		r.TemplateInput.Pipeline, r.TemplateInput.Application)

	// 12. post hook
	c.postHook(ctx, hook.CreateCluster, ret)
	return ret, nil
}

func (c *controller) UpdateCluster(ctx context.Context, clusterID uint,
	r *UpdateClusterRequest) (_ *GetClusterResponse, err error) {
	const op = "cluster controller: update cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. get cluster from db
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. get application that this cluster belongs to
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. get environmentRegion for this cluster
	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	var templateRelease string
	if r.Template == nil || r.Template.Release == "" {
		templateRelease = cluster.TemplateRelease
	} else {
		templateRelease = r.Template.Release
	}

	// 4. if templateInput is not empty, validate templateInput and update templateInput in git repo
	var applicationJSONBlob, pipelineJSONBlob map[string]interface{}
	if r.TemplateInput != nil {
		applicationJSONBlob = r.TemplateInput.Application
		pipelineJSONBlob = r.TemplateInput.Pipeline
		// validate template input
		tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, templateRelease)
		if err != nil {
			return nil, errors.E(op, err)
		}

		renderValues := make(map[string]string)
		clusterIDStr := strconv.FormatUint(uint64(clusterID), 10)
		renderValues[templateschema.ClusterIDKey] = clusterIDStr
		if err := c.validateTemplateInput(ctx,
			cluster.Template, templateRelease, r.TemplateInput, renderValues); err != nil {
			return nil, errors.E(op, http.StatusBadRequest,
				errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
		}
		// update cluster in git repo
		if err := c.clusterGitRepo.UpdateCluster(ctx, &gitrepo.UpdateClusterParams{
			BaseParams: &gitrepo.BaseParams{
				Cluster:             cluster.Name,
				PipelineJSONBlob:    r.TemplateInput.Pipeline,
				ApplicationJSONBlob: r.TemplateInput.Application,
				TemplateRelease:     tr,
				Application:         application,
				Environment:         er.EnvironmentName,
			},
		}); err != nil {
			return nil, errors.E(op, err)
		}
	} else {
		files, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, errors.E(op, err)
		}
		applicationJSONBlob = files.ApplicationJSONBlob
		pipelineJSONBlob = files.PipelineJSONBlob
	}

	// 5. update cluster in db
	clusterModel := r.toClusterModel(cluster, templateRelease)
	clusterModel.UpdatedBy = currentUser.GetID()
	cluster, err = c.clusterMgr.UpdateByID(ctx, clusterID, clusterModel)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 6. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	return ofClusterModel(application, cluster, er, fullPath,
		pipelineJSONBlob, applicationJSONBlob), nil
}

func (c *controller) GetClusterByName(ctx context.Context,
	clusterName string) (_ *GetClusterByNameResponse, err error) {
	const op = "cluster controller: get cluster by name"
	l := wlog.Start(ctx, op)
	defer func() {
		// errors like ClusterNotFound are logged with info level
		if err != nil && errors.Status(err) == http.StatusNotFound {
			log.WithFiled(ctx, "op",
				op).WithField("duration", l.GetDuration().String()).Info(wlog.ByErr(err))
		} else {
			l.Stop(func() string { return wlog.ByErr(err) })
		}
	}()

	// 1. get cluster
	cluster, err := c.clusterMgr.GetByName(ctx, clusterName)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if cluster == nil {
		return nil, errors.E(op, http.StatusNotFound, errors.ErrorCode("ClusterNotFound"))
	}

	// 2. get application
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	return &GetClusterByNameResponse{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: cluster.Description,
		Template: &Template{
			Name:    cluster.Template,
			Release: cluster.TemplateRelease,
		},
		Git: &Git{
			URL:       cluster.GitURL,
			Subfolder: cluster.GitSubfolder,
			Branch:    cluster.GitBranch,
		},
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
		FullPath:  fullPath,
	}, nil
}

// DeleteCluster TODO(gjq): failed to delete cluster, give user a alert.
// TODO(gjq): add a deleting tag for cluster
func (c *controller) DeleteCluster(ctx context.Context, clusterID uint) (err error) {
	const op = "cluster controller: delete cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// get some relevant models
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return errors.E(op, err)
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return errors.E(op, err)
	}

	// 0. set cluster status
	cluster.Status = clustercommon.StatusDeleting
	cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return errors.E(op, err)
	}

	// delete cluster asynchronously, if any error occurs, ignore and return
	go func() {
		// should use a new context
		rid, err := requestid.FromContext(ctx)
		if err != nil {
			log.Errorf(ctx, "failed to get request id from context")
		}
		db, err := orm.FromContext(ctx)
		if err != nil {
			log.Errorf(ctx, "failed to get db from context")
			return
		}
		ctx := log.WithContext(context.Background(), rid)
		ctx = orm.NewContext(ctx, db)

		// 1. delete cluster in cd system
		if err := c.cd.DeleteCluster(ctx, &cd.DeleteClusterParams{
			Environment: er.EnvironmentName,
			Cluster:     cluster.Name,
		}); err != nil {
			log.Errorf(ctx, "failed to delete cluster: %v in cd system, err: %v", cluster.Name, err)
			return
		}

		// 2. delete harbor repository
		harbor := c.registryFty.GetByHarborConfig(ctx, &registry.HarborConfig{
			Server:          regionEntity.Harbor.Server,
			Token:           regionEntity.Harbor.Token,
			PreheatPolicyID: regionEntity.Harbor.PreheatPolicyID,
		})

		if err := harbor.DeleteRepository(ctx, application.Name, cluster.Name); err != nil {
			// log error, not return here, delete harbor repository failed has no effect
			log.Errorf(ctx, "failed to delete harbor repository: %v, err: %v", cluster.Name, err)
		}

		// 3. delete cluster in git repo
		if err := c.clusterGitRepo.DeleteCluster(ctx, application.Name, cluster.Name, cluster.ID); err != nil {
			log.Errorf(ctx, "failed to delete cluster: %v in git repo, err: %v", cluster.Name, err)
		}

		// 4. delete cluster in db
		if err := c.clusterMgr.DeleteByID(ctx, clusterID); err != nil {
			log.Errorf(ctx, "failed to delete cluster: %v in db, err: %v", cluster.Name, err)
		}

		// 5. post hook
		c.postHook(ctx, hook.DeleteCluster, cluster.Name)
	}()

	return nil
}

// FreeCluster to set cluster free
func (c *controller) FreeCluster(ctx context.Context, clusterID uint) (err error) {
	const op = "cluster controller: free cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// get some relevant models
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return errors.E(op, err)
	}

	// 1. set cluster status
	cluster.Status = clustercommon.StatusFreeing
	cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return errors.E(op, err)
	}

	// delete cluster asynchronously, if any error occurs, ignore and return
	go func() {
		// should use a new context
		rid, err := requestid.FromContext(ctx)
		if err != nil {
			log.Errorf(ctx, "failed to get request id from context")
		}
		db, err := orm.FromContext(ctx)
		if err != nil {
			log.Errorf(ctx, "failed to get db from context")
			return
		}
		ctx := log.WithContext(context.Background(), rid)
		ctx = orm.NewContext(ctx, db)

		// 2. delete cluster in cd system
		if err := c.cd.DeleteCluster(ctx, &cd.DeleteClusterParams{
			Environment: er.EnvironmentName,
			Cluster:     cluster.Name,
		}); err != nil {
			log.Errorf(ctx, "failed to delete cluster: %v in cd system, err: %v", cluster.Name, err)
			return
		}

		// 3. set cluster status
		cluster.Status = clustercommon.StatusFreed
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			log.Errorf(ctx, "failed to update cluster: %v, err: %v", cluster.Name, err)
			return
		}
	}()

	return nil
}

// validateCreate validate for create cluster
func (c *controller) validateCreate(r *CreateClusterRequest) error {
	if err := validateClusterName(r.Name); err != nil {
		return err
	}
	if r.Git == nil || r.Git.Branch == "" {
		return fmt.Errorf("git branch cannot be empty")
	}
	if r.TemplateInput != nil && r.TemplateInput.Application == nil {
		return fmt.Errorf("application config for template cannot be empty")
	}
	if r.TemplateInput != nil && r.TemplateInput.Pipeline == nil {
		return fmt.Errorf("pipeline config for template cannot be empty")
	}
	return nil
}

// validateTemplateInput validate templateInput is valid for template schema
func (c *controller) validateTemplateInput(ctx context.Context,
	template, release string, templateInput *TemplateInput, templateSchemaRenderVal map[string]string) error {
	schema, err := c.templateSchemaGetter.GetTemplateSchema(ctx, template, release, templateSchemaRenderVal)
	if err != nil {
		return err
	}
	if err := jsonschema.Validate(schema.Application.JSONSchema, templateInput.Application); err != nil {
		return err
	}
	return jsonschema.Validate(schema.Pipeline.JSONSchema, templateInput.Pipeline)
}

// validateClusterName validate cluster name
// 1. name length must be less than 53
// 2. name must match pattern ^(([a-z][-a-z0-9]*)?[a-z0-9])?$
// 3. name must start with application name
func validateClusterName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 53 {
		return fmt.Errorf("name must not exceed 53 characters")
	}

	// cannot start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		return fmt.Errorf("name cannot start with a digit")
	}

	pattern := `^(([a-z][-a-z0-9]*)?[a-z0-9])?$`
	r := regexp.MustCompile(pattern)
	if !r.MatchString(name) {
		return fmt.Errorf("invalid cluster name, regex used for validation is %v", pattern)
	}

	return nil
}
