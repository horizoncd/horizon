// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"regexp"
	"time"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/cd"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	cmodels "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/cluster/registry"
	collectionmodels "github.com/horizoncd/horizon/pkg/collection/models"
	emvregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/util/jsonschema"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/mergemap"
	"github.com/horizoncd/horizon/pkg/util/permission"
	"github.com/horizoncd/horizon/pkg/util/wlog"

	"github.com/Masterminds/sprig"
	kyaml "sigs.k8s.io/yaml"
)

func (c *controller) List(ctx context.Context, query *q.Query) ([]*ListClusterWithFullResponse, int, error) {
	applicationIDs := make([]uint, 0)

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	currentUserID := currentUser.GetID()

	// get current user
	if query != nil &&
		query.Keywords != nil &&
		query.Keywords[common.ClusterQueryByUser] != nil {
		if userID, ok := query.Keywords[common.ClusterQueryByUser].(uint); ok {
			if err := permission.OnlySelfAndAdmin(ctx, userID); err != nil {
				return nil, 0, err
			}
			currentUserID = userID
			// get groups authorized to current user
			groupIDs, err := c.memberManager.ListResourceOfMemberInfo(ctx, membermodels.TypeGroup, userID)
			if err != nil {
				return nil, 0,
					perror.WithMessage(err, "failed to list group resource of current user")
			}

			// get these groups' subGroups
			subGroups, err := c.groupManager.GetSubGroupsByGroupIDs(ctx, groupIDs)
			if err != nil {
				return nil, 0, perror.WithMessage(err, "failed to get groups")
			}

			subGroupIDs := make([]uint, 0)
			for _, group := range subGroups {
				subGroupIDs = append(subGroupIDs, group.ID)
			}

			// list applications of these subGroups
			applications, err := c.applicationMgr.GetByGroupIDs(ctx, subGroupIDs)
			if err != nil {
				return nil, 0, perror.WithMessage(err, "failed to get applications")
			}

			for _, application := range applications {
				applicationIDs = append(applicationIDs, application.ID)
			}

			// get applications authorized to current user
			authorizedApplicationIDs, err := c.memberManager.ListResourceOfMemberInfo(ctx,
				membermodels.TypeApplication, userID)
			if err != nil {
				return nil, 0,
					perror.WithMessage(err, "failed to list application resource of current user")
			}

			// all applicationIDs, including:
			// (1) applications under the authorized groups
			// (2) authorized applications directly
			applicationIDs = append(applicationIDs, authorizedApplicationIDs...)
		}
	}

	count, clusters, err := c.clusterMgr.List(ctx, query, applicationIDs...)
	if err != nil {
		return nil, 0,
			perror.WithMessage(err, "failed to list user clusters")
	}

	responses, err := c.getFullResponsesWithRegion(ctx, clusters)
	if err != nil {
		return nil, 0, err
	}

	if _, ok := query.Keywords[common.ClusterQueryWithFavorite]; ok {
		err = c.addIsFavoriteForClusters(ctx, currentUserID, responses)
		if err != nil {
			return nil, 0, err
		}
	}
	return responses, count, nil
}

func (c *controller) ListByApplication(ctx context.Context,
	query *q.Query) (_ int, _ []*ListClusterWithFullResponse, err error) {
	const op = "cluster controller: list cluster"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return 0, nil, err
	}
	currentUserID := currentUser.GetID()

	count, clustersWithEnvAndRegion, err := c.clusterMgr.List(ctx, query)
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

	clusters := ofClustersWithEnvRegionTags(clustersWithEnvAndRegion, tags)
	for _, cluster := range clusters {
		cluster.Git.HTTPURL, err = c.commitGetter.GetHTTPLink(cluster.Git.GitURL)
		if err != nil {
			return 0, nil, err
		}
	}

	responses := make([]*ListClusterWithFullResponse, 0, len(clusters))
	for _, cluster := range clusters {
		responses = append(responses, &ListClusterWithFullResponse{
			ListClusterResponse: cluster,
		})
	}

	if _, ok := query.Keywords[common.ClusterQueryWithFavorite]; ok {
		err = c.addIsFavoriteForClusters(ctx, currentUserID, responses)
		if err != nil {
			return 0, nil, err
		}
	}

	return count, responses, nil
}

func (c *controller) getFullResponses(ctx context.Context,
	clusters []*cmodels.Cluster) ([]*ListClusterWithFullResponse, error) {
	// get applications
	var applicationIDs []uint
	for _, cluster := range clusters {
		applicationIDs = append(applicationIDs, cluster.ApplicationID)
	}
	applicationMap, err := c.applicationSvc.GetByIDs(ctx, applicationIDs)
	if err != nil {
		return nil, err
	}

	responses := make([]*ListClusterWithFullResponse, 0)

	// 3. convert and add full path, full name
	for _, cluster := range clusters {
		application, exist := applicationMap[cluster.ApplicationID]
		if !exist {
			continue
		}
		fullPath := fmt.Sprintf("%v/%v", application.FullPath, cluster.Name)
		fullName := fmt.Sprintf("%v/%v", application.FullName, cluster.Name)
		response := ofCluster(cluster)
		response.Git.HTTPURL, err = c.commitGetter.GetHTTPLink(response.Git.GitURL)
		if err != nil {
			return nil, err
		}
		responses = append(responses, &ListClusterWithFullResponse{
			response,
			nil,
			fullName,
			fullPath,
		})
	}
	return responses, nil
}

func (c *controller) getFullResponsesWithRegion(ctx context.Context,
	clustersWithRegion []*cmodels.ClusterWithRegion) ([]*ListClusterWithFullResponse, error) {
	clusters := make([]*cmodels.Cluster, 0, len(clustersWithRegion))
	for _, clusterWithRegion := range clustersWithRegion {
		clusters = append(clusters, clusterWithRegion.Cluster)
	}

	responses, err := c.getFullResponses(ctx, clusters)
	if err != nil {
		return nil, err
	}

	for i := range responses {
		responses[i].Scope.RegionDisplayName = clustersWithRegion[i].RegionDisplayName
	}
	return responses, nil
}

func (c *controller) ListClusterWithExpiry(ctx context.Context,
	query *q.Query) ([]*ListClusterWithExpiryResponse, error) {
	const op = "cluster controller: list clusters with expiry"
	defer wlog.Start(ctx, op).StopPrint()
	clusterList, err := c.clusterMgr.ListClusterWithExpiry(ctx, query)
	return ofClusterWithExpiry(clusterList), err
}

func (c *controller) clusterWillExpireIn(ctx context.Context, cluster *cmodels.Cluster) (*uint, error) {
	if cluster.ExpireSeconds == 0 {
		return nil, nil
	}

	latestPipelinerun, err := c.getLatestPipelinerunByClusterID(ctx, cluster.ID)
	if err != nil {
		return nil, err
	}
	if latestPipelinerun == nil {
		return nil, nil
	}

	return willExpireIn(cluster.ExpireSeconds, cluster.UpdatedAt, latestPipelinerun.UpdatedAt), nil
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

	// 7. get tags
	tags, err := c.tagMgr.ListByResourceTypeID(ctx, common.ResourceCluster, clusterID)
	if err != nil {
		return nil, err
	}

	// 8. transfer model
	clusterResp := ofClusterModel(application, cluster, fullPath, envValue.Namespace,
		clusterFiles.PipelineJSONBlob, clusterFiles.ApplicationJSONBlob, tags...)

	// 9. get latest deployed commit
	latestPR, err := c.prMgr.PipelineRun.GetLatestSuccessByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if latestPR != nil {
		clusterResp.LatestDeployedCommit = latestPR.GitCommit
		clusterResp.Image = latestPR.ImageURL
	}

	// 10. get createdBy and updatedBy users
	userMap, err := c.userManager.GetUserMapByIDs(ctx, []uint{cluster.CreatedBy, cluster.UpdatedBy})
	if err != nil {
		return nil, err
	}
	clusterResp.CreatedBy = toUser(getUserFromMap(cluster.CreatedBy, userMap))
	clusterResp.UpdatedBy = toUser(getUserFromMap(cluster.UpdatedBy, userMap))
	if cluster.Status != common.ClusterStatusFreed &&
		latestPR != nil {
		clusterResp.TTLInSeconds, _ = c.clusterWillExpireIn(ctx, cluster)
	}

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
	oneMap := make(map[string]interface{})
	var err error
	for _, clusterValueFile := range clusterValueFiles {
		if clusterValueFile.Content != nil {
			if content, ok := clusterValueFile.Content[templateName]; ok {
				// if content is empty or {}, continue
				if contentMap, ok := content.(map[string]interface{}); ok && len(contentMap) > 0 {
					oneMap, err = mergemap.Merge(oneMap, contentMap)
					if err != nil {
						return nil, perror.Wrapf(herrors.ErrParamInvalid, "merge map error, err = %s", err.Error())
					}
				}
			}
		}
	}

	var addValuePrefixDocMap = make(map[interface{}]interface{})
	addValuePrefixDocMap[_valuePrefix] = oneMap
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

func (c *controller) CreateCluster(ctx context.Context, applicationID uint, environment,
	region string, r *CreateClusterRequest, mergePatch bool) (_ *GetClusterResponse, err error) {
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

	if err := c.customizeTemplateInfo(ctx, r, application, environment, mergePatch); err != nil {
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

	// transfer expireTime to expireSeconds and verify environment.
	// expireTime's format is e.g. "300ms", "-1.5h" or "2h45m".
	expireSeconds, err := c.toExpireSeconds(ctx, r.ExpireTime, environment)
	if err != nil {
		return nil, err
	}

	// 5. get templateRelease
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, r.Template.Name, r.Template.Release)
	if err != nil {
		return nil, err
	}

	// 6. create cluster, after created, params.Cluster is the newest cluster
	cluster, tags := r.toClusterModel(application, er, expireSeconds)
	cluster.Status = common.ClusterStatusCreating

	if err := tagmanager.ValidateUpsert(tags); err != nil {
		return nil, err
	}

	// 7. create cluster in db
	cluster, err = c.clusterMgr.Create(ctx, cluster, tags, r.ExtraMembers)
	if err != nil {
		return nil, err
	}

	// TODO: refactor by asynchronous task, and notify api callers to adapt
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
	cluster.Status = common.ClusterStatusEmpty
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
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}

	ret := ofClusterModel(application, cluster, fullPath, envValue.Namespace,
		r.TemplateInput.Pipeline, r.TemplateInput.Application)

	// 11. record event
	c.recordClusterEvent(ctx, ret.ID, eventmodels.ClusterCreated)
	c.recordMemberCreatedEvent(ctx, ret.ID)
	return ret, nil
}

func (c *controller) UpdateCluster(ctx context.Context, clusterID uint,
	r *UpdateClusterRequest, mergePatch bool) (_ *GetClusterResponse, err error) {
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

	if r.ExpireTime != "" {
		expireSeconds, err := c.toExpireSeconds(ctx, r.ExpireTime, cluster.EnvironmentName)
		if err != nil {
			return nil, err
		}
		cluster.ExpireSeconds = expireSeconds
	}

	// can only update environment/region when the cluster has been freed
	if cluster.Status == common.ClusterStatusFreed && r.Environment != "" && r.Region != "" {
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

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, templateRelease)
	if err != nil {
		return nil, err
	}

	clusterModel, tags := r.toClusterModel(cluster, templateRelease, er)

	// 4. if templateInput is not empty, validate templateInput and update templateInput in git repo
	if r.TemplateInput != nil {
		// merge cluster config and request config
		// merge patch allows users to pass only some fields
		if mergePatch {
			files, err := c.clusterGitRepo.GetCluster(ctx, application.Name,
				cluster.Name, cluster.Template)
			if err != nil {
				return nil, err
			}

			r.TemplateInput.Application, err = mergemap.Merge(files.ApplicationJSONBlob,
				r.TemplateInput.Application)
			if err != nil {
				return nil, err
			}

			r.TemplateInput.Pipeline, err = mergemap.Merge(files.PipelineJSONBlob,
				r.TemplateInput.Pipeline)
			if err != nil {
				return nil, err
			}
		}

		// validate template input
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
		files, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, tr.ChartName)
		if err != nil {
			return nil, err
		}
		r.TemplateInput = &TemplateInput{
			Application: files.ApplicationJSONBlob,
			Pipeline:    files.PipelineJSONBlob,
		}
	}

	// 5. update cluster in db
	// todo: atomicity
	cluster, err = c.clusterMgr.UpdateByID(ctx, clusterID, clusterModel)
	if err != nil {
		return nil, err
	}

	// 7. update cluster tags
	tagsInDB, err := c.tagMgr.ListByResourceTypeID(ctx, common.ResourceCluster, clusterID)
	if err != nil {
		return nil, err
	}
	if r.Tags != nil && !tagmodels.Tags(tags).Eq(tagsInDB) {
		if err := c.clusterGitRepo.UpdateTags(ctx, application.Name, cluster.Name, cluster.Template, tags); err != nil {
			return nil, err
		}
		if err := c.tagMgr.UpsertByResourceTypeID(ctx, common.ResourceCluster, clusterID, r.Tags); err != nil {
			return nil, err
		}
	}

	// 6. record event
	c.recordClusterEvent(ctx, cluster.ID, eventmodels.ClusterUpdated)

	// 7. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	// 8. get namespace
	if namespaceChanged {
		envValue, err = c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, err
		}
	}

	return ofClusterModel(application, cluster, fullPath, envValue.Namespace,
		r.TemplateInput.Pipeline, r.TemplateInput.Application, tags...), nil
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
	cluster.Status = common.ClusterStatusDeleting
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

		// 2. delete image
		rg, err := c.registryFty.GetRegistryByConfig(newctx, &registry.Config{
			Server:             regionEntity.Registry.Server,
			Token:              regionEntity.Registry.Token,
			InsecureSkipVerify: regionEntity.Registry.InsecureSkipTLSVerify,
			Kind:               regionEntity.Registry.Kind,
			Path:               regionEntity.Registry.Path,
		})

		if err != nil {
			log.Errorf(newctx, "failed to get registry by config: err = %v", err)
		}

		if rg != nil {
			if err = rg.DeleteImage(newctx, application.Name, cluster.Name); err != nil {
				// log error, not return here, delete image failed has no effect
				log.Errorf(newctx, "failed to delete image: %v, err: %v", cluster.Name, err)
			}
		}

		if hard {
			// delete member
			if err := c.memberManager.HardDeleteMemberByResourceTypeID(ctx,
				string(membermodels.TypeApplicationCluster), clusterID); err != nil {
				log.Errorf(newctx, "failed to delete members of cluster: %v, err: %v", cluster.Name, err)
			}
			// delete pipelinerun
			if err := c.prMgr.PipelineRun.DeleteByClusterID(ctx, clusterID); err != nil {
				log.Errorf(newctx, "failed to delete pipelineruns of cluster: %v, err: %v", cluster.Name, err)
			}
			// delete tag
			if err := c.tagMgr.UpsertByResourceTypeID(ctx, common.ResourceCluster, clusterID, nil); err != nil {
				log.Errorf(newctx, "failed to delete tags of cluster: %v, err: %v", cluster.Name, err)
			}
			// delete gitrepo
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

		// 5. record event
		c.recordClusterEvent(newctx, clusterID, eventmodels.ClusterDeleted)
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
	if cluster.Status == common.ClusterStatusFreeing {
		log.Warningf(ctx, "failed to free cluster: %v, cluster status: %v", cluster.Name, cluster.Status)
		return nil
	}

	// 1. set cluster status
	cluster.Status = common.ClusterStatusFreeing
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
			cluster.Status = common.ClusterStatusFreed
			if err != nil {
				cluster.Status = ""
				_, err = c.clusterMgr.UpdateByID(newctx, cluster.ID, cluster)
				if err != nil {
					log.Errorf(newctx, "failed to update cluster: %v, err: %v", cluster.Name, err)
				}
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
		cluster.Status = common.ClusterStatusFreed
		_, err = c.clusterMgr.UpdateByID(newctx, cluster.ID, cluster)
		if err != nil {
			log.Errorf(newctx, "failed to update cluster: %v, err: %v", cluster.Name, err)
			return
		}

		// 4. create event
		c.recordClusterEvent(newctx, clusterID, eventmodels.ClusterFreed)
	}()

	return nil
}

func (c *controller) toExpireSeconds(ctx context.Context, expireTime string, environment string) (uint, error) {
	expireSeconds := uint(0)
	if expireTime != "" {
		duration, err := time.ParseDuration(expireTime)
		if err != nil {
			log.Errorf(ctx, "failed to parse expireTime, err: %v", err.Error())
			return 0, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
		expireSeconds = uint(duration.Seconds())
		if !c.autoFreeSvc.WhetherSupported(environment) && expireSeconds > 0 {
			log.Warningf(ctx, "%v environment dose not support auto-free, but expireSeconds are %v",
				environment, expireSeconds)
			expireSeconds = 0
		}
	}
	return expireSeconds, nil
}

func (c *controller) customizeTemplateInfo(ctx context.Context, r *CreateClusterRequest,
	application *models.Application, environment string, mergePatch bool) error {
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

	appGitRepo, err := c.applicationGitRepo.GetApplication(ctx, application.Name, environment)
	if err != nil {
		return err
	}
	pipelineJSONBlob := appGitRepo.BuildConf
	applicationJSONBlob := appGitRepo.TemplateConf
	if r.TemplateInput == nil {
		r.TemplateInput = &TemplateInput{}
		r.TemplateInput.Application = applicationJSONBlob
		r.TemplateInput.Pipeline = pipelineJSONBlob
	} else if mergePatch {
		// merge patch allows users to pass only some fields
		applicationJSONBlob, err := mergemap.Merge(applicationJSONBlob,
			r.TemplateInput.Application)
		if err != nil {
			return err
		}
		pipelineJSONBlob, err := mergemap.Merge(pipelineJSONBlob,
			r.TemplateInput.Pipeline)
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
	return status == common.ClusterStatusCreating || status == common.ClusterStatusDeleting
}

var (
	favoriteTrue  = true
	favoriteFalse = false
)

func (c *controller) addIsFavoriteForClusters(ctx context.Context,
	userID uint, clusters []*ListClusterWithFullResponse) error {
	ids := make([]uint, 0, len(clusters))
	for i := range clusters {
		ids = append(ids, clusters[i].ID)
	}
	collections, err := c.collectionManager.List(ctx, userID, common.ResourceCluster, ids)
	if err != nil {
		return err
	}
	m := map[uint]collectionmodels.Collection{}
	for _, collection := range collections {
		m[collection.ResourceID] = collection
	}

	for i := range clusters {
		if _, ok := m[clusters[i].ID]; ok {
			clusters[i].IsFavorite = &favoriteTrue
		} else {
			clusters[i].IsFavorite = &favoriteFalse
		}
	}
	return nil
}
