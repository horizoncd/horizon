package cluster

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/core/common"
	controllertag "g.hz.netease.com/horizon/core/controller/tag"
	herrors "g.hz.netease.com/horizon/core/errors"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/util/jsonschema"
	"g.hz.netease.com/horizon/pkg/util/mergemap"

	codemodels "g.hz.netease.com/horizon/pkg/cluster/code"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

func (c *controller) CreateClusterV2(ctx context.Context, applicationID uint, environment,
	region string, r *CreateClusterRequestV2, mergePatch bool) (*CreateClusterResponseV2, error) {
	const op = "cluster controller: create cluster v2"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. check exist
	exists, err := c.clusterMgr.CheckClusterExists(ctx, r.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, perror.Wrap(herrors.ErrNameConflict,
			"a cluster with the same name already exists, please do not create it again")
	}

	// 2. validate create req
	if err := validateClusterName(r.Name); err != nil {
		return nil, err
	}

	// 3. get application
	application, err := c.applicationMgr.GetByID(ctx, applicationID)
	if err != nil {
		return nil, err
	}

	// 4. customize buildTemplateInfo and do validate
	buildTemplateInfo, err := c.customizeBuildTemplateInfo(ctx, r, application, environment, mergePatch)
	if err != nil {
		return nil, err
	}
	if err := buildTemplateInfo.Validate(ctx, c.templateSchemaGetter); err != nil {
		return nil, err
	}

	// 5. get environment and region
	envEntity, err := c.envRegionMgr.GetByEnvironmentAndRegion(ctx, environment, region)
	if err != nil {
		return nil, err
	}
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, region)
	if err != nil {
		return nil, err
	} else {
		if regionEntity.Disabled {
			return nil, perror.Wrap(herrors.ErrDisabled,
				"the region which is disabled cannot be used to create a cluster")
		}
	}

	// 6 get templateRelease
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, r.TemplateInfo.Name, r.TemplateInfo.Release)
	if err != nil {
		return nil, err
	}

	// 7. customize db infos
	cluster, tags := r.toClusterModel(application, envEntity, buildTemplateInfo)

	// 8. update db and tags
	clusterResp, err := c.clusterMgr.Create(ctx, cluster, tags, r.ExtraMembers)
	if err != nil {
		return nil, err
	}

	// 9. create git repo
	err = c.clusterGitRepo.CreateCluster(ctx, &gitrepo.CreateClusterParams{
		BaseParams: &gitrepo.BaseParams{
			ClusterID:           clusterResp.ID,
			Cluster:             clusterResp.Name,
			PipelineJSONBlob:    r.BuildConfig,
			ApplicationJSONBlob: r.TemplateConfig,
			TemplateRelease:     tr,
			Application:         application,
			Environment:         environment,
			RegionEntity:        regionEntity,
			Version:             gitrepo.VersionV2,
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
	updateClusterResp, err := c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
	if err != nil {
		return nil, err
	}

	// 10. get full path
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

	// 11. customize response
	return &CreateClusterResponseV2{
		ID:            updateClusterResp.ID,
		FullPath:      fullPath,
		ApplicationID: applicationID,
		CreatedAt:     updateClusterResp.CreatedAt,
		UpdatedAt:     updateClusterResp.UpdatedAt,
	}, nil
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

	// 4. gen fullpath
	group, err := c.groupSvc.GetChildByID(ctx, application.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v/%v", group.FullPath, application.Name, cluster.Name)

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

	// 9. get createdBy and updatedBy users
	userMap, err := c.userManager.GetUserMapByIDs(ctx, []uint{cluster.CreatedBy, cluster.UpdatedBy})

	getResp := &GetClusterResponseV2{
		ID:          cluster.ID,
		Name:        cluster.Name,
		Description: cluster.Description,
		// TODO: currently cluster not support different priority with application
		Priority: string(application.Priority),
		Scope: &Scope{
			Environment:       cluster.EnvironmentName,
			Region:            cluster.RegionName,
			RegionDisplayName: regionEntity.DisplayName,
		},
		FullPath:        fullPath,
		ApplicationName: application.Name,
		ApplicationID:   application.ID,
		Tags: func() []controllertag.Tag {
			retTags := make([]controllertag.Tag, 0, len(tags))
			for _, tag := range retTags {
				retTags = append(retTags, controllertag.Tag{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
			return retTags
		}(),
		Git: func() *codemodels.Git {
			if cluster.GitURL != "" {
				return nil
			}
			return codemodels.NewGit(cluster.GitURL, cluster.GitSubfolder,
				cluster.GitRefType, cluster.GitRef)
		}(),
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
	return getResp, nil
}

func (c *controller) UpdateClusterV2(ctx context.Context, clusterID uint,
	r *UpdateClusterRequest, mergePatch bool) error {
	return nil
}

type BuildTemplateInfo struct {
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
	// Manifest       map[string]interface{}   `json:"manifest"`
}

func (info *BuildTemplateInfo) Validate(ctx context.Context, trGetter templateschema.Getter) error {
	// templateSchemaRenderVal := make(map[string]string)
	templateSchemaRenderVal := make(map[string]string)
	// TODO (remove it, currently some template need it)
	templateSchemaRenderVal["resourceType"] = "cluster"
	schema, err := trGetter.GetTemplateSchema(ctx, info.TemplateInfo.Name,
		info.TemplateInfo.Release, templateSchemaRenderVal)
	if err != nil {
		return err
	}
	return jsonschema.Validate(schema.Application.JSONSchema, info.TemplateConfig, false)
}

func (c *controller) customizeBuildTemplateInfo(ctx context.Context, r *CreateClusterRequestV2,
	application *appmodels.Application, environment string, mergePatch bool) (*BuildTemplateInfo, error) {

	buildTemplateInfo := &BuildTemplateInfo{}
	appGitRepo, err := c.applicationGitRepo.GetApplication(ctx, application.Name, environment)
	if err != nil {
		return nil, err
	}
	if r.BuildConfig == nil {
		if !mergePatch {
			buildTemplateInfo.BuildConfig = appGitRepo.BuildConf
		} else {
			buildTemplateInfo.BuildConfig, err = mergemap.Merge(appGitRepo.BuildConf, r.BuildConfig)
			if err != nil {
				return nil, err
			}
		}
	} else {
		buildTemplateInfo.BuildConfig = r.BuildConfig
	}

	if r.TemplateInfo == nil && r.TemplateConfig == nil {
		buildTemplateInfo.TemplateInfo = &codemodels.TemplateInfo{
			Name:    application.Template,
			Release: application.TemplateRelease,
		}
		if !mergePatch {
			buildTemplateInfo.TemplateConfig = appGitRepo.TemplateConf
		} else {
			buildTemplateInfo.TemplateConfig, err = mergemap.Merge(appGitRepo.BuildConf, r.BuildConfig)
			if err != nil {
				return nil, err
			}
		}
	} else if r.TemplateInfo != nil && r.TemplateConfig != nil {
		buildTemplateInfo.TemplateInfo = r.TemplateInfo
		buildTemplateInfo.TemplateConfig = r.TemplateConfig
	} else {
		return nil, perror.Wrap(herrors.ErrParamInvalid, "TemplateInfo or TemplateConfig nil")
	}
	return buildTemplateInfo, nil
}
