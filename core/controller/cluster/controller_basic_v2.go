package cluster

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/application/models"
	codemodels "g.hz.netease.com/horizon/pkg/cluster/code"
	perror "g.hz.netease.com/horizon/pkg/errors"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
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
	if err := c.ValidateCreateV2(r); err != nil {
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
	if err := buildTemplateInfo.Validate(ctx, c.templateReleaseMgr); err != nil {
		return nil, err
	}

	// 5. update db and tags

	// 6. create git repo

	// 7. customize response

	return nil, nil
}

type BuildTemplateInfo struct {
	BuildConfig    map[string]interface{}   `json:"buildConfig"`
	TemplateInfo   *codemodels.TemplateInfo `json:"templateInfo"`
	TemplateConfig map[string]interface{}   `json:"templateConfig"`
	Manifest       map[string]interface{}   `json:"manifest"`
}

func (info *BuildTemplateInfo) Validate(ctx context.Context, templateManger trmanager.Manager) error {
	// templateSchemaRenderVal := make(map[string]string)
	return nil
}

func (c *controller) customizeBuildTemplateInfo(ctx context.Context, r *CreateClusterRequestV2,
	application *models.Application, environment string, mergePatch bool) (*BuildTemplateInfo, error) {
	return nil, nil
}

func (c *controller) ValidateCreateV2(r *CreateClusterRequestV2) error {
	return nil
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
	_, err = c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *controller) UpdateClusterV2(ctx context.Context, clusterID uint,
	r *UpdateClusterRequest, mergePatch bool) error {
	return nil
}
