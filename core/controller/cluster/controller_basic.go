package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	codemodels "g.hz.netease.com/horizon/pkg/cluster/code"
	clustercommon "g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/registry"
	emvregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	tagmanager "g.hz.netease.com/horizon/pkg/tag/manager"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	"g.hz.netease.com/horizon/pkg/util/jsonschema"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	"github.com/Masterminds/sprig"
	"github.com/go-yaml/yaml"
	kyaml "sigs.k8s.io/yaml"
)

func (c *controller) ListCluster(ctx context.Context, applicationID uint, environments []string,
	filter string, query *q.Query, ts []tagmodels.TagSelector) (_ int, _ []*ListClusterResponse, err error) {
	const op = "cluster controller: list cluster"
	defer wlog.Start(ctx, op).StopPrint()

	count, clustersWithEnvAndRegion, err := c.clusterMgr.ListByApplicationEnvsTags(ctx,
		applicationID, environments, filter, query, ts)
	if err != nil {
		return 0, nil, err
	}

	clusterIDs := []uint{}
	for _, c := range clustersWithEnvAndRegion {
		clusterIDs = append(clusterIDs, c.ID)
	}

	tags, err := c.tagMgr.ListByResourceTypeIDs(ctx, common.ResourceCluster, clusterIDs, false)
	if err != nil {
		return 0, nil, err
	}

	return count, ofClustersWithEnvRegionTags(clustersWithEnvAndRegion, tags), nil
}

func (c *controller) ListClusterByNameFuzzily(ctx context.Context, environment,
	filter string, query *q.Query) (count int, listClusterWithFullResp []*ListClusterWithFullResponse, err error) {
	const op = "cluster controller: list cluster by name fuzzily"
	defer wlog.Start(ctx, op).StopPrint()

	listClusterWithFullResp = []*ListClusterWithFullResponse{}
	// 1. get clusters
	count, clustersWithEnvAndRegion, err := c.clusterMgr.ListByNameFuzzily(ctx,
		environment, filter, query)
	if err != nil {
		return 0, nil, err
	}

	// 2. get applications
	var applicationIDs []uint
	for _, cluster := range clustersWithEnvAndRegion {
		applicationIDs = append(applicationIDs, cluster.ApplicationID)
	}
	applicationMap, err := c.applicationSvc.GetByIDs(ctx, applicationIDs)
	if err != nil {
		return 0, nil, err
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

func (c *controller) ListUserClusterByNameFuzzily(ctx context.Context, environment,
	filter string, query *q.Query) (count int, resp []*ListClusterWithFullResponse, err error) {
	// get current user
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return 0, nil, perror.WithMessage(err, "no user in context")
	}

	// get groups authorized to current user
	groupIDs, err := c.memberManager.ListResourceOfMemberInfo(ctx, membermodels.TypeGroup, currentUser.GetID())
	if err != nil {
		return 0, nil,
			perror.WithMessage(err, "failed to list group resource of current user")
	}

	// get these groups' subGroups
	subGroups, err := c.groupManager.GetSubGroupsByGroupIDs(ctx, groupIDs)
	if err != nil {
		return 0, nil, perror.WithMessage(err, "failed to get groups")
	}

	subGroupIDs := make([]uint, 0)
	for _, group := range subGroups {
		subGroupIDs = append(subGroupIDs, group.ID)
	}

	// list applications of these subGroups
	applications, err := c.applicationMgr.GetByGroupIDs(ctx, subGroupIDs)
	if err != nil {
		return 0, nil, perror.WithMessage(err, "failed to get applications")
	}

	applicationIDs := make([]uint, 0)
	for _, application := range applications {
		applicationIDs = append(applicationIDs, application.ID)
	}

	// get applications authorized to current user
	authorizedApplicationIDs, err := c.memberManager.ListResourceOfMemberInfo(ctx,
		membermodels.TypeApplication, currentUser.GetID())
	if err != nil {
		return 0, nil,
			perror.WithMessage(err, "failed to list application resource of current user")
	}

	// all applicationIDs, including:
	// (1) applications under the authorized groups
	// (2) authorized applications directly
	applicationIDs = append(applicationIDs, authorizedApplicationIDs...)

	count, clusters, err := c.clusterMgr.ListUserAuthorizedByNameFuzzily(ctx, environment,
		filter, applicationIDs, currentUser.GetID(), query)
	if err != nil {
		return 0, nil,
			perror.WithMessage(err, "failed to list user clusters")
	}

	// 2. get applications
	clusterApplicationIDs := make([]uint, 0)
	for _, cluster := range clusters {
		clusterApplicationIDs = append(clusterApplicationIDs, cluster.ApplicationID)
	}
	applicationMap, err := c.applicationSvc.GetByIDs(ctx, clusterApplicationIDs)
	if err != nil {
		return 0, nil,
			perror.WithMessage(err, "failed to list application for clusters")
	}

	resp = make([]*ListClusterWithFullResponse, 0)
	// 3. convert and add full path, full name
	for _, cluster := range clusters {
		application, exist := applicationMap[cluster.ApplicationID]
		if !exist {
			continue
		}
		fullPath := fmt.Sprintf("%v/%v", application.FullPath, cluster.Name)
		fullName := fmt.Sprintf("%v/%v", application.FullName, cluster.Name)
		resp = append(resp, &ListClusterWithFullResponse{
			ofClusterWithEnvAndRegion(cluster),
			fullName,
			fullPath,
		})
	}

	return count, resp, nil
}

func (c *controller) GetCluster(ctx context.Context, clusterID uint) (_ *GetClusterResponse, err error) {
	const op = "cluster controller: get cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get cluster from db
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 2. get application
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 4. get files in git repo
	clusterFiles := &gitrepo.ClusterFiles{}
	if !isClusterStatusUnstable(cluster.Status) {
		clusterFiles, err = c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, err
		}
	}

	// 5. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	// 6. get namespace
	envValue := &gitrepo.EnvValue{}
	if !isClusterStatusUnstable(cluster.Status) {
		envValue, err = c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, err
		}
	}

	// 7. transfer model
	clusterResp := ofClusterModel(application, cluster, fullPath, envValue.Namespace,
		clusterFiles.PipelineJSONBlob, clusterFiles.ApplicationJSONBlob)

	// 8. get latest deployed commit
	latestPR, err := c.pipelinerunMgr.GetLatestSuccessByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if latestPR != nil {
		clusterResp.LatestDeployedCommit = latestPR.GitCommit
		clusterResp.Image = latestPR.ImageURL
	}

	// 9. get createdBy and updatedBy users
	userMap, err := c.userManager.GetUserMapByIDs(ctx, []uint{cluster.CreatedBy, cluster.UpdatedBy})
	if err != nil {
		return nil, err
	}
	clusterResp.CreatedBy = toUser(getUserFromMap(cluster.CreatedBy, userMap))
	clusterResp.UpdatedBy = toUser(getUserFromMap(cluster.UpdatedBy, userMap))

	return clusterResp, nil
}

func (c *controller) GetClusterOutput(ctx context.Context, clusterID uint) (_ interface{}, err error) {
	const op = "cluster controller: get cluster output"
	defer wlog.Start(ctx, op).StopPrint()
	// 1. get cluster from db
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 2. get application
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. get output in template
	outputStr, err := c.outputGetter.GetTemplateOutPut(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	if outputStr == "" {
		return nil, nil
	}

	// 4. get files in  git repo
	clusterFiles, err := c.clusterGitRepo.GetClusterValueFiles(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}
	if len(clusterFiles) == 0 {
		return nil, nil
	}

	log.Debugf(ctx, "clusterFiles = %+v, outputStr = %+v", clusterFiles, outputStr)

	// 5. reader output in template and return
	outputRenderJSONObject, err := RenderOutputObject(outputStr, cluster.Template, clusterFiles...)
	if err != nil {
		return nil, err
	}

	return outputRenderJSONObject, nil
}

const (
	_valuePrefix = "Values"
)

func RenderOutputObject(outPutStr, templateName string,
	clusterValueFiles ...gitrepo.ClusterValueFile) (interface{}, error) {
	// remove the  template prefix level, add Value prefix(as helm) and merge to one doc
	var oneDoc string
	for _, clusterValueFile := range clusterValueFiles {
		if clusterValueFile.Content != nil {
			if content, ok := clusterValueFile.Content[templateName]; ok {
				// if content is empty or {}, continue
				if contentMap, ok := content.(map[interface{}]interface{}); !ok || len(contentMap) == 0 {
					continue
				}
				binaryContent, err := yaml.Marshal(content)
				if err != nil {
					return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
				}
				oneDoc += string(binaryContent) + "\n"
			}
		}
	}
	var oneDocMap map[interface{}]interface{}
	err := yaml.Unmarshal([]byte(oneDoc), &oneDocMap)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "RenderOutputObject yaml Unmarshal  error, err  = %s", err.Error())
	}

	var addValuePrefixDocMap = make(map[interface{}]interface{})
	addValuePrefixDocMap[_valuePrefix] = oneDocMap
	var b bytes.Buffer
	doTemplate := template.Must(template.New("").Funcs(sprig.HtmlFuncMap()).Parse(outPutStr))
	err = doTemplate.ExecuteTemplate(&b, "", addValuePrefixDocMap)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "RenderOutputObject template error, err  = %s", err.Error())
	}

	var retJSONObject interface{}
	jsonBytes, err := kyaml.YAMLToJSON(b.Bytes())
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "RenderOutputObject YAMLToJSON error, err  = %s", err.Error())
	}
	err = json.Unmarshal(jsonBytes, &retJSONObject)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "RenderOutputObject json Unmarshal error, err  = %s", err.Error())
	}
	return retJSONObject, nil
}

func (c *controller) CreateCluster(ctx context.Context, applicationID uint,
	environment, region string, r *CreateClusterRequest) (_ *GetClusterResponse, err error) {
	const op = "cluster controller: create cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get application
	application, err := c.applicationMgr.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// 2. validate
	exists, err := c.clusterMgr.CheckClusterExists(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, perror.Wrap(herrors.ErrNameConflict,
			"a cluster with the same name already exists, please do not create it again")
	}
	if err := c.validateCreate(r); err != nil {
		return nil, err
	}

	users := make([]string, 0, len(r.ExtraMembers))
	for member := range r.ExtraMembers {
		users = append(users, member)
	}

	err = c.userSvc.CheckUsersExists(ctx, users)
	if err != nil {
		return nil, err
	}

	if err := c.customizeTemplateInfo(ctx, r, application, environment); err != nil {
		return nil, err
	}
	if err := c.validateTemplateInput(ctx, r.Template.Name,
		r.Template.Release, r.TemplateInput, nil); err != nil {
		return nil, err
	}

	// 3. get environmentRegion
	er, err := c.envRegionMgr.GetByEnvironmentAndRegion(ctx, environment, region)
	if err != nil {
		return nil, err
	}

	// 4. get regionEntity
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, region)
	if err != nil {
		return nil, err
	}
	if regionEntity.Disabled {
		return nil, perror.Wrap(herrors.ErrDisabled,
			"the region which is disabled cannot be used to create a cluster")
	}

	// 5. get templateRelease
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, r.Template.Name, r.Template.Release)
	if err != nil {
		return nil, err
	}

	// 6. create cluster, after created, params.Cluster is the newest cluster
	cluster, tags := r.toClusterModel(application, er)
	cluster.Status = clustercommon.StatusCreating

	if err := tagmanager.ValidateUpsert(tags); err != nil {
		return nil, err
	}

	// 7. create cluster in db
	cluster, err = c.clusterMgr.Create(ctx, cluster, tags, r.ExtraMembers)
	if err != nil {
		return nil, err
	}

	// TODO: refactor by asynchronous task, and notify overmind/faas to adapt
	// 8. create cluster in git repo
	err = c.clusterGitRepo.CreateCluster(ctx, &gitrepo.CreateClusterParams{
		BaseParams: &gitrepo.BaseParams{
			ClusterID:           cluster.ID,
			Cluster:             cluster.Name,
			PipelineJSONBlob:    r.TemplateInput.Pipeline,
			ApplicationJSONBlob: r.TemplateInput.Application,
			TemplateRelease:     tr,
			Application:         application,
			Environment:         environment,
			RegionEntity:        regionEntity,
			Namespace:           r.Namespace,
		},
		Tags:  tags,
		Image: r.Image,
	})
	if err != nil {
		// Prevent errors like "project has already been taken" caused by automatic retries due to api timeouts
		if deleteErr := c.clusterGitRepo.DeleteCluster(ctx, application.Name, cluster.Name, cluster.ID); deleteErr != nil {
			if _, ok := perror.Cause(deleteErr).(*herrors.HorizonErrNotFound); !ok {
				err = perror.WithMessage(err, deleteErr.Error())
			}
		}
		if deleteErr := c.clusterMgr.DeleteByID(ctx, cluster.ID); deleteErr != nil {
			err = perror.WithMessage(err, deleteErr.Error())
		}
		return nil, err
	}
	cluster.Status = clustercommon.StatusEmpty
	cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return nil, err
	}

	// 9. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	// 10. get namespace
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}

	ret := ofClusterModel(application, cluster, fullPath, envValue.Namespace,
		r.TemplateInput.Pipeline, r.TemplateInput.Application)

	// 11. post hook
	c.postHook(ctx, hook.CreateCluster, ret)
	return ret, nil
}

func (c *controller) UpdateCluster(ctx context.Context, clusterID uint,
	r *UpdateClusterRequest) (_ *GetClusterResponse, err error) {
	const op = "cluster controller: update cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get cluster from db
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	// 2. get application that this cluster belongs to
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. get environmentRegion/namespace for this cluster
	var (
		er               *emvregionmodels.EnvironmentRegion
		regionEntity     *regionmodels.RegionEntity
		namespace        string
		namespaceChanged bool
	)

	// can only update environment/region when the cluster has been freed
	if cluster.Status == clustercommon.StatusFreed && r.Environment != "" && r.Region != "" {
		er, err = c.envRegionMgr.GetByEnvironmentAndRegion(ctx, r.Environment, r.Region)
		if err != nil {
			return nil, err
		}
		regionEntity, err = c.regionMgr.GetRegionEntity(ctx, er.RegionName)
		if err != nil {
			return nil, err
		}
	} else {
		er = &emvregionmodels.EnvironmentRegion{
			EnvironmentName: cluster.EnvironmentName,
			RegionName:      cluster.RegionName,
		}
	}

	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}

	// if environment has not changed, keep the namespace unchanged (for cloudnative app)
	if er.EnvironmentName == cluster.EnvironmentName {
		namespace = envValue.Namespace
	} else {
		namespaceChanged = true
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
			return nil, err
		}
		renderValues, err := c.getRenderValueFromTag(ctx, clusterID)
		if err != nil {
			return nil, err
		}
		if err := c.validateTemplateInput(ctx,
			cluster.Template, templateRelease, r.TemplateInput, renderValues); err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"request body validate err: %v", err)
		}
		// update cluster in git repo
		if err := c.clusterGitRepo.UpdateCluster(ctx, &gitrepo.UpdateClusterParams{
			BaseParams: &gitrepo.BaseParams{
				ClusterID:           cluster.ID,
				Cluster:             cluster.Name,
				PipelineJSONBlob:    r.TemplateInput.Pipeline,
				ApplicationJSONBlob: r.TemplateInput.Application,
				TemplateRelease:     tr,
				Application:         application,
				Environment:         er.EnvironmentName,
				RegionEntity:        regionEntity,
				Namespace:           namespace,
			},
		}); err != nil {
			return nil, err
		}
	} else {
		files, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, err
		}
		applicationJSONBlob = files.ApplicationJSONBlob
		pipelineJSONBlob = files.PipelineJSONBlob
	}

	// 5. update cluster in db
	clusterModel := r.toClusterModel(cluster, templateRelease, er)
	// todo: atomicity
	cluster, err = c.clusterMgr.UpdateByID(ctx, clusterID, clusterModel)
	if err != nil {
		return nil, err
	}

	// 6. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	// 7. get namespace
	if namespaceChanged {
		envValue, err = c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, err
		}
	}

	return ofClusterModel(application, cluster, fullPath, envValue.Namespace,
		pipelineJSONBlob, applicationJSONBlob), nil
}

func (c *controller) GetClusterByName(ctx context.Context,
	clusterName string) (_ *GetClusterByNameResponse, err error) {
	const op = "cluster controller: get cluster by name"
	wlog.Start(ctx, op).StopPrint()

	// 1. get cluster
	cluster, err := c.clusterMgr.GetByName(ctx, clusterName)
	if err != nil {
		return nil, err
	}
	if cluster == nil {
		return nil, err
	}

	// 2. get application
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
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
		Git: codemodels.NewGit(
			cluster.GitURL,
			cluster.GitSubfolder,
			cluster.GitRefType,
			cluster.GitRef,
		),
		CreatedAt: cluster.CreatedAt,
		UpdatedAt: cluster.UpdatedAt,
		FullPath:  fullPath,
	}, nil
}

// DeleteCluster TODO(gjq): failed to delete cluster, give user a alert.
// TODO(gjq): add a deleting tag for cluster
func (c *controller) DeleteCluster(ctx context.Context, clusterID uint, hard bool) (err error) {
	const op = "cluster controller: delete cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// get some relevant models
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return err
	}

	// 0. set cluster status
	cluster.Status = clustercommon.StatusDeleting
	cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return err
	}

	// should use a new context
	rid, err := requestid.FromContext(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get request id from context")
	}
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return
	}

	newctx := log.WithContext(context.Background(), rid)
	newctx = common.WithContext(newctx, currentUser)
	// delete cluster asynchronously, if any error occurs, ignore and return
	go func() {
		var err error
		defer func() {
			if err != nil {
				cluster.Status = ""
				_, err = c.clusterMgr.UpdateByID(newctx, cluster.ID, cluster)
				if err != nil {
					log.Errorf(newctx, "failed to update cluster: %v, err: %v", cluster.Name, err)
				}
			}
		}()

		// 1. delete cluster in cd system
		if err = c.cd.DeleteCluster(newctx, &cd.DeleteClusterParams{
			Environment: cluster.EnvironmentName,
			Cluster:     cluster.Name,
		}); err != nil {
			log.Errorf(newctx, "failed to delete cluster: %v in cd system, err: %v", cluster.Name, err)
			return
		}

		// 2. delete harbor repository
		harbor := c.registryFty.GetByHarborConfig(newctx, &registry.HarborConfig{
			Server:          regionEntity.Harbor.Server,
			Token:           regionEntity.Harbor.Token,
			PreheatPolicyID: regionEntity.Harbor.PreheatPolicyID,
		})

		if err = harbor.DeleteRepository(newctx, application.Name, cluster.Name); err != nil {
			// log error, not return here, delete harbor repository failed has no effect
			log.Errorf(newctx, "failed to delete harbor repository: %v, err: %v", cluster.Name, err)
		}

		if hard {
			if err := c.pipelinerunMgr.DeleteByClusterID(ctx, clusterID); err != nil {
				log.Errorf(newctx, "failed to delete pipelineruns of cluster: %v, err: %v", cluster.Name, err)
			}

			if err := c.pipelineMgr.DeleteByClusterName(ctx, cluster.Name); err != nil {
				log.Errorf(newctx, "failed to delete pipelineruns of cluster: %v, err: %v", cluster.Name, err)
			}

			if err := c.tagMgr.UpsertByResourceTypeID(ctx, common.ResourceCluster, clusterID, nil); err != nil {
				log.Errorf(newctx, "failed to delete tags of cluster: %v, err: %v", cluster.Name, err)
			}

			err = c.clusterGitRepo.HardDeleteCluster(newctx, application.Name, cluster.Name)
			if err != nil {
				if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
					log.Errorf(newctx, "failed to delete cluster: %v in git repo, err: %v", cluster.Name, err)
				}
			}
		} else {
			if err = c.clusterGitRepo.DeleteCluster(newctx, application.Name, cluster.Name, cluster.ID); err != nil {
				log.Errorf(newctx, "failed to delete cluster: %v in git repo, err: %v", cluster.Name, err)
			}
		}

		// 4. delete cluster in db
		if err = c.clusterMgr.DeleteByID(newctx, clusterID); err != nil {
			log.Errorf(newctx, "failed to delete cluster: %v in db, err: %v", cluster.Name, err)
		}

		// 5. post hook
		c.postHook(newctx, hook.DeleteCluster, cluster.Name)
	}()

	return nil
}

// FreeCluster to set cluster free
func (c *controller) FreeCluster(ctx context.Context, clusterID uint) (err error) {
	const op = "cluster controller: free cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// get some relevant models
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}

	// 1. set cluster status
	cluster.Status = clustercommon.StatusFreeing
	cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return err
	}

	// should use a new context
	rid, err := requestid.FromContext(ctx)
	if err != nil {
		log.Errorf(ctx, "failed to get request id from context")
	}
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return
	}
	newctx := log.WithContext(context.Background(), rid)
	newctx = common.WithContext(newctx, currentUser)
	// delete cluster asynchronously, if any error occurs, ignore and return
	go func() {
		var err error
		defer func() {
			cluster.Status = clustercommon.StatusFreed
			if err != nil {
				cluster.Status = ""
			}
			_, err = c.clusterMgr.UpdateByID(newctx, cluster.ID, cluster)
			if err != nil {
				log.Errorf(newctx, "failed to update cluster: %v, err: %v", cluster.Name, err)
				return
			}
		}()

		// 2. delete cluster in cd system
		if err = c.cd.DeleteCluster(newctx, &cd.DeleteClusterParams{
			Environment: cluster.EnvironmentName,
			Cluster:     cluster.Name,
		}); err != nil {
			log.Errorf(newctx, "failed to delete cluster: %v in cd system, err: %v", cluster.Name, err)
			return
		}

		// 3. set cluster status
		cluster.Status = clustercommon.StatusFreed
		_, err = c.clusterMgr.UpdateByID(newctx, cluster.ID, cluster)
		if err != nil {
			log.Errorf(newctx, "failed to update cluster: %v, err: %v", cluster.Name, err)
			return
		}
	}()

	return nil
}

func (c *controller) customizeTemplateInfo(ctx context.Context,
	r *CreateClusterRequest, application *models.Application, environment string) error {
	// 1. if template is empty, set it with application's template
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

	// 2. if templateInput is empty, set it with application's env template
	if r.TemplateInput == nil {
		pipelineJSONBlob, applicationJSONBlob, err := c.applicationGitRepo.
			GetApplicationEnvTemplate(ctx, application.Name, environment)
		if err != nil {
			return err
		}
		r.TemplateInput = &TemplateInput{}
		r.TemplateInput.Application = applicationJSONBlob
		r.TemplateInput.Pipeline = pipelineJSONBlob
	}
	return nil
}

func (c *controller) getRenderValueFromTag(ctx context.Context, clusterID uint) (map[string]string, error) {
	tags, err := c.schemaTagManager.ListByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	renderValues := make(map[string]string)
	for _, tag := range tags {
		renderValues[tag.Key] = tag.Value
	}
	return renderValues, nil
}

// validateCreate validate for create cluster
func (c *controller) validateCreate(r *CreateClusterRequest) error {
	if err := validateClusterName(r.Name); err != nil {
		return err
	}
	if r.Git == nil || r.Git.Ref() == "" || r.Git.RefType() == "" {
		return perror.Wrap(herrors.ErrParamInvalid, "git ref cannot be empty")
	}
	if r.TemplateInput != nil && r.TemplateInput.Application == nil {
		return perror.Wrap(herrors.ErrParamInvalid, "application config for template cannot be empty")
	}
	if r.TemplateInput != nil && r.TemplateInput.Pipeline == nil {
		return perror.Wrap(herrors.ErrParamInvalid, "pipeline config for template cannot be empty")
	}
	return nil
}

// validateTemplateInput validate templateInput is valid for template schema
func (c *controller) validateTemplateInput(ctx context.Context,
	template, release string, templateInput *TemplateInput, templateSchemaRenderVal map[string]string) error {
	if templateSchemaRenderVal == nil {
		templateSchemaRenderVal = make(map[string]string)
	}
	// TODO (remove it, currently some template need it)
	templateSchemaRenderVal["resourceType"] = "cluster"
	schema, err := c.templateSchemaGetter.GetTemplateSchema(ctx, template, release, templateSchemaRenderVal)
	if err != nil {
		return err
	}
	if err := jsonschema.Validate(schema.Application.JSONSchema, templateInput.Application, false); err != nil {
		return err
	}
	return jsonschema.Validate(schema.Pipeline.JSONSchema, templateInput.Pipeline, true)
}

// validateClusterName validate cluster name
// 1. name length must be less than 53
// 2. name must match pattern ^(([a-z][-a-z0-9]*)?[a-z0-9])?$
// 3. name must start with application name
func validateClusterName(name string) error {
	if len(name) == 0 {
		return perror.Wrap(herrors.ErrParamInvalid, "name cannot be empty")
	}

	if len(name) > 53 {
		return perror.Wrap(herrors.ErrParamInvalid, "name must not exceed 53 characters")
	}

	// cannot start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		return perror.Wrap(herrors.ErrParamInvalid, "name cannot start with a digit")
	}

	pattern := `^(([a-z][-a-z0-9]*)?[a-z0-9])?$`
	r := regexp.MustCompile(pattern)
	if !r.MatchString(name) {
		return perror.Wrapf(herrors.ErrParamInvalid, "invalid cluster name, regex used for validation is %v", pattern)
	}

	return nil
}

// isUnstableStatus judge if status is Creating or Deleting
func isClusterStatusUnstable(status string) bool {
	return status == clustercommon.StatusCreating || status == clustercommon.StatusDeleting
}
