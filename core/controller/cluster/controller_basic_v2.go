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
	"context"
	"fmt"
	"time"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/build"
	herrors "github.com/horizoncd/horizon/core/errors"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/cd"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	collectionmodels "github.com/horizoncd/horizon/pkg/collection/models"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/git"
	prmodels "github.com/horizoncd/horizon/pkg/pr/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/templaterelease/models"
	templateschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
	"github.com/horizoncd/horizon/pkg/util/jsonschema"
	"github.com/horizoncd/horizon/pkg/util/mergemap"
	"github.com/horizoncd/horizon/pkg/util/validate"

	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	perror "github.com/horizoncd/horizon/pkg/errors"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"

	"github.com/horizoncd/horizon/pkg/util/wlog"
)

func (c *controller) GetClusterStatusV2(ctx context.Context, clusterID uint) (*StatusResponseV2, error) {
	const op = "cluster controller: get cluster status v2"
	defer wlog.Start(ctx, op).StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	resp := &StatusResponseV2{}
	// status in db has higher priority
	if cluster.Status != common.ClusterStatusEmpty {
		resp.Status = cluster.Status
	}

	params := &cd.GetClusterStateV2Params{
		Application:  application.Name,
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		RegionEntity: regionEntity,
	}

	// If cluster not found on argo, check the cluster is freed or has not been published yet,
	// or the cluster will be "notFound". If response's status field has not been set,
	// set it with status from argocd.
	cdStatus, err := c.cd.GetClusterState(ctx, params)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return nil, err
		}
		// for not deployed -- free or not published
		if resp.Status == "" {
			resp.Status = _notFound
		}
	} else {
		// resp has not been set
		if resp.Status == "" {
			resp.Status = cdStatus.Status
		}
	}

	return resp, nil
}

func (c *controller) CreateClusterV2(ctx context.Context,
	params *CreateClusterParamsV2) (*CreateClusterResponseV2, error) {
	const op = "cluster controller: create cluster v2"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. check exist
	exists, err := c.clusterMgr.CheckClusterExists(ctx, params.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, perror.Wrap(herrors.ErrNameConflict,
			"a cluster with the same name already exists, please do not create it again")
	}

	// 2. validate create req
	if err := validateClusterName(params.Name); err != nil {
		return nil, err
	}
	if params.Git != nil && params.Git.URL != "" {
		if err := validate.CheckGitURL(params.Git.URL); err != nil {
			return nil, err
		}
	}
	if params.Image != nil {
		if err := validate.CheckImageURL(*params.Image); err != nil {
			return nil, err
		}
	}

	// 3. get application
	application, err := c.applicationMgr.GetByID(ctx, params.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 4. customize buildTemplateInfo and do validate
	buildTemplateInfo, err := c.customizeCreateReqBuildTemplateInfo(ctx, params, application)
	if err != nil {
		return nil, err
	}
	if err := buildTemplateInfo.Validate(ctx,
		c.templateSchemaGetter, nil, c.buildSchema); err != nil {
		return nil, err
	}

	// 5. get environment and region
	envEntity, err := c.envRegionMgr.GetByEnvironmentAndRegion(ctx,
		params.Environment, params.Region)
	if err != nil {
		return nil, err
	}
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, params.Region)
	if err != nil {
		return nil, err
	}
	if regionEntity.Disabled {
		return nil, perror.Wrap(herrors.ErrDisabled,
			"the region which is disabled cannot be used to create a cluster")
	}

	// 6. transfer expireTime to expireSeconds and verify environment.
	expireSeconds, err := c.toExpireSeconds(ctx, params.ExpireTime, params.Environment)
	if err != nil {
		return nil, err
	}

	// 7. get template and templateRelease
	template, err := c.templateMgr.GetByName(ctx, buildTemplateInfo.TemplateInfo.Name)
	if err != nil {
		return nil, err
	}
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx,
		buildTemplateInfo.TemplateInfo.Name, buildTemplateInfo.TemplateInfo.Release)
	if err != nil {
		return nil, err
	}

	// 8. customize db infos
	cluster, tags := params.toClusterModel(application,
		envEntity, buildTemplateInfo, template, expireSeconds)

	// 9. update db and tags
	clusterResp, err := c.clusterMgr.Create(ctx, cluster, tags, params.ExtraMembers)
	if err != nil {
		return nil, err
	}

	// 10. create git repo
	err = c.clusterGitRepo.CreateCluster(ctx, &gitrepo.CreateClusterParams{
		BaseParams: &gitrepo.BaseParams{
			ClusterID:           clusterResp.ID,
			Cluster:             clusterResp.Name,
			PipelineJSONBlob:    buildTemplateInfo.BuildConfig,
			ApplicationJSONBlob: buildTemplateInfo.TemplateConfig,
			TemplateRelease:     tr,
			Application:         application,
			Environment:         params.Environment,
			RegionEntity:        regionEntity,
			Version:             common.MetaVersion2,
		},
		Tags: tags,
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
	updateClusterResp, err := c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return nil, err
	}

	// 11. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}

	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	ret := &CreateClusterResponseV2{
		ID:   updateClusterResp.ID,
		Name: updateClusterResp.Name,
		Application: &Application{
			ID:   application.ID,
			Name: application.Name,
		},
		Scope: &Scope{
			Environment: cluster.EnvironmentName,
			Region:      cluster.RegionName,
		},
		FullPath:      fullPath,
		ApplicationID: application.ID,
		CreatedAt:     updateClusterResp.CreatedAt,
		UpdatedAt:     updateClusterResp.UpdatedAt,
	}

	// 12. record event
	c.eventSvc.CreateEventIgnoreError(ctx, common.ResourceCluster, ret.ID,
		eventmodels.ClusterCreated, nil)
	c.eventSvc.RecordMemberCreatedEvent(ctx, common.ResourceCluster, ret.ID)
	// 13. customize response
	return ret, nil
}

func (c *controller) GetClusterV2(ctx context.Context, clusterID uint) (*GetClusterResponseV2, error) {
	const op = "cluster controller: get cluster v2"
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

	// 3. get region entity
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	// 4. gen fullPath
	fullPath, err := func() (string, error) {
		group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name), nil
	}()
	if err != nil {
		return nil, err
	}

	// 5. get tags
	tags, err := c.tagMgr.ListByResourceTypeID(ctx, common.ResourceCluster, cluster.ID)
	if err != nil {
		return nil, err
	}

	// 6. get GitRepo
	clusterGitRepoFile, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, err
	}

	// 7. get createdBy and updatedBy users
	userMap, err := c.userManager.GetUserMapByIDs(ctx, []uint{cluster.CreatedBy, cluster.UpdatedBy})
	if err != nil {
		return nil, err
	}
	getResp := &GetClusterResponseV2{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: cluster.Description,
		// TODO: currently it's not allowed that the cluster has different priority with related application.
		Priority: string(application.Priority),
		Scope: &Scope{
			Environment:       cluster.EnvironmentName,
			Region:            cluster.RegionName,
			RegionDisplayName: regionEntity.DisplayName,
		},
		ExpireTime: func() string {
			expireTime := ""
			if cluster.ExpireSeconds > 0 {
				expireTime = time.Duration(cluster.ExpireSeconds * 1e9).String()
			}
			return expireTime
		}(),
		FullPath:        fullPath,
		ApplicationName: application.Name,
		ApplicationID:   application.ID,
		Tags:            tagmodels.Tags(tags).IntoTagsBasic(),
		Git: func() *codemodels.Git {
			if cluster.GitURL == "" {
				return nil
			}
			return codemodels.NewGit(cluster.GitURL, cluster.GitSubfolder,
				cluster.GitRefType, cluster.GitRef)
		}(),
		Image:       cluster.Image,
		BuildConfig: clusterGitRepoFile.PipelineJSONBlob,
		TemplateInfo: func() *codemodels.TemplateInfo {
			if cluster.Template == "" {
				return nil
			}
			return &codemodels.TemplateInfo{
				Name:    cluster.Template,
				Release: cluster.TemplateRelease,
			}
		}(),
		TemplateConfig: clusterGitRepoFile.ApplicationJSONBlob,
		Manifest:       clusterGitRepoFile.Manifest,
		Status:         cluster.Status,
		CreatedAt:      cluster.CreatedAt,
		UpdatedAt:      cluster.UpdatedAt,
		CreatedBy:      toUser(getUserFromMap(cluster.CreatedBy, userMap)),
		UpdatedBy:      toUser(getUserFromMap(cluster.UpdatedBy, userMap)),
	}

	// 8. get latest deployed commit
	latestPR, err := c.prMgr.PipelineRun.GetLatestSuccessByClusterID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if cluster.Status != common.ClusterStatusFreed &&
		latestPR != nil {
		getResp.TTLInSeconds, _ = c.clusterWillExpireIn(ctx, cluster)
	}
	return getResp, nil
}

func (c *controller) UpdateClusterV2(ctx context.Context, clusterID uint,
	r *UpdateClusterRequestV2, mergePatch bool) error {
	const op = "cluster controller: update cluster v2"
	defer wlog.Start(ctx, op).StopPrint()

	// validate request
	if r.Git != nil && r.Git.URL != "" {
		if err := validate.CheckGitURL(r.Git.URL); err != nil {
			return err
		}
	}
	if r.Image != nil && *r.Image != "" {
		if err := validate.CheckImageURL(*r.Image); err != nil {
			return err
		}
	}

	// 1. get cluster and application from db
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return err
	}

	// 2. check if we should update region and env
	var regionEntity *regionmodels.RegionEntity
	environmentName := cluster.EnvironmentName
	regionName := cluster.RegionName
	if cluster.Status == common.ClusterStatusFreed &&
		(r.Environment != nil && *r.Environment != environmentName) ||
		(r.Region != nil && *r.Region != regionName) {
		if r.Environment != nil {
			environmentName = *r.Environment
		}
		if r.Region != nil {
			regionName = *r.Region
		}
		regionEntity, err = c.regionMgr.GetRegionEntity(ctx, regionName)
		if err != nil {
			return err
		}
		_, err = c.envRegionMgr.GetByEnvironmentAndRegion(ctx, environmentName, regionName)
		if err != nil {
			return err
		}
	}

	// 3. check and transfer ExpireTime
	expireSeconds := cluster.ExpireSeconds
	if r.ExpireTime != "" {
		expireSeconds, err = c.toExpireSeconds(ctx, r.ExpireTime, environmentName)
		if err != nil {
			return err
		}
	}

	// 4. customize template\build\template infos
	templateInfo, templateRelease, err := func() (*codemodels.TemplateInfo, *models.TemplateRelease, error) {
		var templateInfo *codemodels.TemplateInfo
		if r.TemplateInfo == nil {
			templateInfo = &codemodels.TemplateInfo{
				Name:    cluster.Template,
				Release: cluster.TemplateRelease,
			}
		} else {
			templateInfo = r.TemplateInfo
		}
		tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx,
			templateInfo.Name, templateInfo.Release)
		if err != nil {
			return nil, nil, err
		}
		return templateInfo, tr, nil
	}()
	if err != nil {
		return err
	}

	buildConfig, templateConfig, err := func() (map[string]interface{}, map[string]interface{}, error) {
		if r.BuildConfig == nil && r.TemplateConfig == nil {
			return nil, nil, nil
		}
		files, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, nil, err
		}
		if files.Manifest == nil {
			return nil, nil, perror.Wrapf(herrors.ErrParamInvalid, "git repo  %s not support v2 interface",
				cluster.Name)
		}

		buildConfig := r.BuildConfig
		templateConfig := r.TemplateConfig
		if r.BuildConfig != nil && mergePatch {
			buildConfig, err = mergemap.Merge(files.PipelineJSONBlob, r.BuildConfig)
			if err != nil {
				return nil, nil, err
			}
		}
		if r.TemplateConfig != nil && mergePatch {
			templateConfig, err = mergemap.Merge(files.ApplicationJSONBlob, r.TemplateConfig)
			if err != nil {
				return nil, nil, err
			}
		}
		return buildConfig, templateConfig, nil
	}()
	if err != nil {
		return err
	}

	// 5. validate update Request
	err = func() error {
		renderValues, err := c.getRenderValueFromTag(ctx, clusterID)
		if err != nil {
			return err
		}
		info := BuildTemplateInfo{
			BuildConfig:    buildConfig,
			TemplateInfo:   templateInfo,
			TemplateConfig: templateConfig,
		}
		return info.Validate(ctx, c.templateSchemaGetter, renderValues, c.buildSchema)
	}()
	if err != nil {
		return err
	}

	// 6. update in git repo
	if err = c.clusterGitRepo.UpdateCluster(ctx, &gitrepo.UpdateClusterParams{
		BaseParams: &gitrepo.BaseParams{
			ClusterID:           cluster.ID,
			Cluster:             cluster.Name,
			PipelineJSONBlob:    buildConfig,
			ApplicationJSONBlob: templateConfig,
			TemplateRelease:     templateRelease,
			Application:         application,
			Environment:         environmentName,
			RegionEntity:        regionEntity,
			Version:             common.MetaVersion2,
		}}); err != nil {
		return err
	}

	// 7. record event
	c.eventSvc.CreateEventIgnoreError(ctx, common.ResourceCluster, cluster.ID,
		eventmodels.ClusterUpdated, nil)

	// 8. update cluster in db
	clusterModel, tags := r.toClusterModel(cluster, expireSeconds, environmentName,
		regionName, templateInfo.Name, templateInfo.Release)
	_, err = c.clusterMgr.UpdateByID(ctx, clusterID, clusterModel)
	if err != nil {
		return err
	}

	// 9. update cluster tags
	tagsInDB, err := c.tagMgr.ListByResourceTypeID(ctx, common.ResourceCluster, clusterID)
	if err != nil {
		return err
	}
	if r.Tags != nil && !tagmodels.Tags(tags).Eq(tagsInDB) {
		if err := c.clusterGitRepo.UpdateTags(ctx, application.Name, cluster.Name, cluster.Template, tags); err != nil {
			return err
		}
		if err := c.tagMgr.UpsertByResourceTypeID(ctx, common.ResourceCluster, clusterID, r.Tags); err != nil {
			return err
		}
	}
	return nil
}

type BuildTemplateInfo struct {
	BuildConfig    map[string]interface{}
	TemplateInfo   *codemodels.TemplateInfo
	TemplateConfig map[string]interface{}
}

func (info *BuildTemplateInfo) Validate(ctx context.Context,
	trGetter templateschema.Getter, templateSchemaRenderVal map[string]string, buildSchema *build.Schema) error {
	if templateSchemaRenderVal == nil {
		templateSchemaRenderVal = make(map[string]string)
	}
	// TODO (remove it, currently some template need it)
	templateSchemaRenderVal["resourceType"] = "cluster"
	schema, err := trGetter.GetTemplateSchema(ctx, info.TemplateInfo.Name,
		info.TemplateInfo.Release, templateSchemaRenderVal)
	if err != nil {
		return err
	}
	err = jsonschema.Validate(schema.Application.JSONSchema, info.TemplateConfig, false)
	if err != nil {
		return err
	}

	if buildSchema != nil && info.BuildConfig != nil && len(info.BuildConfig) > 0 {
		err = jsonschema.Validate(buildSchema.JSONSchema, info.BuildConfig, false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *controller) customizeCreateReqBuildTemplateInfo(ctx context.Context, params *CreateClusterParamsV2,
	application *appmodels.Application) (*BuildTemplateInfo, error) {
	buildTemplateInfo := &BuildTemplateInfo{}
	appGitRepoFile, err := c.applicationGitRepo.GetApplication(ctx, application.Name, params.Environment)
	if err != nil {
		return nil, err
	}

	inherit := true

	if params.TemplateInfo != nil {
		buildTemplateInfo.TemplateInfo = params.TemplateInfo
		if params.TemplateInfo.Name != application.Template {
			inherit = false
		}
	} else {
		buildTemplateInfo.TemplateInfo = &codemodels.TemplateInfo{
			Name:    application.Template,
			Release: application.TemplateRelease,
		}
	}

	if inherit {
		// inherit config from application if it's empty in the request
		if params.Git != nil {
			buildTemplateInfo.BuildConfig = appGitRepoFile.BuildConf
		}
		buildTemplateInfo.TemplateConfig = appGitRepoFile.TemplateConf
	}

	if params.TemplateConfig != nil {
		if params.MergePatch {
			buildTemplateInfo.TemplateConfig, err = mergemap.Merge(appGitRepoFile.TemplateConf, params.TemplateConfig)
			if err != nil {
				return nil, err
			}
		} else {
			buildTemplateInfo.TemplateConfig = params.TemplateConfig
		}
	}

	if params.BuildConfig != nil {
		if params.MergePatch {
			buildTemplateInfo.BuildConfig, err = mergemap.Merge(appGitRepoFile.BuildConf, params.BuildConfig)
			if err != nil {
				return nil, err
			}
		} else {
			buildTemplateInfo.BuildConfig = params.BuildConfig
		}
	}
	return buildTemplateInfo, nil
}

func (c *controller) ToggleLikeStatus(ctx context.Context, clusterID uint, like *WhetherLike) (err error) {
	// get current user
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return perror.WithMessage(err, "no user in context")
	}

	if like != nil {
		if like.IsFavorite {
			collection := collectionmodels.Collection{
				ResourceID:   clusterID,
				ResourceType: common.ResourceCluster,
				UserID:       currentUser.GetID(),
			}
			_, err := c.collectionManager.Create(ctx, &collection)
			return err
		}
		_, err := c.collectionManager.DeleteByResource(ctx, currentUser.GetID(), clusterID, common.ResourceCluster)
		return err
	}
	return nil
}

func (c *controller) CreatePipelineRun(ctx context.Context, clusterID uint,
	r *CreatePipelineRunRequest) (*prmodels.PipelineBasic, error) {
	const op = "pipelinerun controller: create pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	pipelineRun, err := c.createPipelineRun(ctx, clusterID, r)
	if err != nil {
		return nil, err
	}

	// if checks is empty, set status to ready
	checks, err := c.prSvc.GetCheckByResource(ctx, clusterID, common.ResourceCluster)
	if err != nil {
		return nil, err
	}
	if len(checks) == 0 {
		pipelineRun.Status = string(prmodels.StatusReady)
	}

	if pipelineRun, err = c.prMgr.PipelineRun.Create(ctx, pipelineRun); err != nil {
		return nil, err
	}

	c.eventSvc.CreateEventIgnoreError(ctx, common.ResourcePipelinerun, pipelineRun.ID,
		eventmodels.PipelinerunCreated, nil)

	firstCanRollbackPipelinerun, err := c.prMgr.PipelineRun.GetFirstCanRollbackPipelinerun(ctx, pipelineRun.ClusterID)
	if err != nil {
		return nil, err
	}
	return c.prSvc.OfPipelineBasic(ctx, pipelineRun, firstCanRollbackPipelinerun)
}

func (c *controller) createPipelineRun(ctx context.Context, clusterID uint,
	r *CreatePipelineRunRequest) (*prmodels.Pipelinerun, error) {
	defer wlog.Start(ctx, "cluster controller: create pipeline run").StopPrint()

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	if r.Action == prmodels.ActionBuildDeploy && cluster.GitURL == "" {
		return nil, herrors.ErrBuildDeployNotSupported
	}

	var (
		title        = r.Title
		action       string
		gitURL       = cluster.GitURL
		gitRefType   = cluster.GitRefType
		gitRef       = cluster.GitRef
		codeCommitID string
		imageURL     = cluster.Image
		rollbackFrom *uint
	)

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}

	configCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, err
	}

	var lastConfigCommitSHA, configCommitSHA = configCommit.Master, configCommit.Gitops

	switch r.Action {
	case prmodels.ActionBuildDeploy:
		action = prmodels.ActionBuildDeploy

		if r.Git != nil {
			if r.Git.Commit != "" {
				gitRefType = codemodels.GitRefTypeCommit
				gitRef = r.Git.Commit
			} else if r.Git.Tag != "" {
				gitRefType = codemodels.GitRefTypeTag
				gitRef = r.Git.Tag
			} else if r.Git.Branch != "" {
				gitRefType = codemodels.GitRefTypeBranch
				gitRef = r.Git.Branch
			}
		}

		commit, err := c.commitGetter.GetCommit(ctx, cluster.GitURL, gitRefType, gitRef)
		if err != nil {
			commit = &git.Commit{
				Message: "commit not found",
				ID:      gitRef,
			}
		}
		codeCommitID = commit.ID

		imageURL = assembleImageURL(regionEntity, application.Name, cluster.Name, gitRef, commit.ID)

	case prmodels.ActionDeploy:
		action = prmodels.ActionDeploy

		clusterFiles, err := c.clusterGitRepo.GetCluster(ctx, application.Name, cluster.Name, cluster.Template)
		if err != nil {
			return nil, err
		}

		if cluster.GitURL != "" {
			err = c.checkAllowDeploy(ctx, application, cluster, clusterFiles, configCommit)
			if err != nil {
				return nil, err
			}

			commit, err := c.commitGetter.GetCommit(ctx, cluster.GitURL, cluster.GitRefType, cluster.GitRef)
			if err == nil {
				codeCommitID = commit.ID
			}
		} else if cluster.Image != "" {
			imageURL, err = getDeployImage(cluster.Image, r.ImageTag)
			if err != nil {
				return nil, err
			}
		}

	case prmodels.ActionRollback:
		title = prmodels.ActionRollback
		action = prmodels.ActionRollback

		// get pipelinerun to rollback, and do some validation
		pipelinerun, err := c.prMgr.PipelineRun.GetByID(ctx, r.PipelinerunID)
		if err != nil {
			return nil, err
		}

		if pipelinerun.Action == prmodels.ActionRestart || pipelinerun.Status != string(prmodels.StatusOK) ||
			pipelinerun.ConfigCommit == "" {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"the pipelinerun with id: %v can not be rolled back", r.PipelinerunID)
		}

		if pipelinerun.ClusterID != cluster.ID {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"the pipelinerun with id: %v is not belongs to cluster: %v", r.PipelinerunID, clusterID)
		}

		gitURL = pipelinerun.GitURL
		gitRefType = pipelinerun.GitRefType
		gitRef = pipelinerun.GitRef
		codeCommitID = pipelinerun.GitCommit
		imageURL = pipelinerun.ImageURL
		rollbackFrom = &pipelinerun.ID
		configCommitSHA = configCommit.Master

	case prmodels.ActionRestart:
		title = prmodels.ActionRestart
		action = prmodels.ActionRestart
		configCommitSHA = configCommit.Master

	default:
		return nil, perror.Wrapf(herrors.ErrParamInvalid, "unsupported action %v", r.Action)
	}

	return &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           action,
		Status:           string(prmodels.StatusPending),
		Title:            title,
		Description:      r.Description,
		GitURL:           gitURL,
		GitRefType:       gitRefType,
		GitRef:           gitRef,
		GitCommit:        codeCommitID,
		ImageURL:         imageURL,
		LastConfigCommit: lastConfigCommitSHA,
		ConfigCommit:     configCommitSHA,
		RollbackFrom:     rollbackFrom,
	}, nil
}
