package deployer

import (
	"context"

	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Params struct {
	Environment         string
	Application         *appmodels.Application
	Cluster             *models.Cluster
	RegionEntity        *regionmodels.RegionEntity
	PipelineJSONBlob    map[string]interface{}
	ApplicationJSONBlob map[string]interface{}
	TemplateRelease     *trmodels.TemplateRelease
}

type Deployer interface {
	CreateCluster(ctx context.Context, params *Params) error
	UpdateCluster(ctx context.Context, params *Params) error
}

type deployer struct {
	cd      cd.CD
	gitRepo gitrepo.ClusterGitRepo
	mgr     manager.Manager
}

func NewDeployer(clusterGitRepo gitrepo.ClusterGitRepo, cd cd.CD) Deployer {
	return &deployer{
		cd:      cd,
		gitRepo: clusterGitRepo,
		mgr:     manager.Mgr,
	}
}

func (d *deployer) CreateCluster(ctx context.Context, params *Params) (err error) {
	const op = "deployer: create cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get helm repo
	helmRepo, err := d.cd.GetHelmRepo(ctx, params.Environment)
	if err != nil {
		return errors.E(op, err)
	}

	// 2. create cluster in git repo
	if err := d.gitRepo.CreateCluster(ctx, &gitrepo.Params{
		Cluster:             params.Cluster.Name,
		HelmRepoURL:         helmRepo,
		Environment:         params.Environment,
		RegionEntity:        params.RegionEntity,
		PipelineJSONBlob:    params.PipelineJSONBlob,
		ApplicationJSONBlob: params.ApplicationJSONBlob,
		TemplateRelease:     params.TemplateRelease,
		Application:         params.Application,
	}); err != nil {
		return errors.E(op, err)
	}

	// 3. create cluster in cd system
	if err := d.cd.CreateCluster(ctx, &cd.Params{
		Application: params.Application.Name,
		Environment: params.Environment,
		Cluster:     params.Cluster.Name,
	}); err != nil {
		return errors.E(op, err)
	}

	// 4. create cluster in db
	cluster, err := d.mgr.Create(ctx, params.Cluster)
	if err != nil {
		return errors.E(op, err)
	}

	params.Cluster = cluster
	return nil
}

func (d *deployer) UpdateCluster(ctx context.Context, params *Params) (err error) {
	const op = "deployer: deploy cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get helm repo
	helmRepo, err := d.cd.GetHelmRepo(ctx, params.Environment)
	if err != nil {
		return errors.E(op, err)
	}

	// 2. update cluster in git repo
	if err := d.gitRepo.UpdateCluster(ctx, &gitrepo.Params{
		Cluster:             params.Cluster.Name,
		HelmRepoURL:         helmRepo,
		Environment:         params.Environment,
		RegionEntity:        params.RegionEntity,
		PipelineJSONBlob:    params.PipelineJSONBlob,
		ApplicationJSONBlob: params.ApplicationJSONBlob,
		TemplateRelease:     params.TemplateRelease,
		Application:         params.Application,
	}); err != nil {
		return errors.E(op, err)
	}

	// 2. update cluster in db
	cluster, err := d.mgr.UpdateByID(ctx, params.Cluster.ID, params.Cluster)
	if err != nil {
		return errors.E(op, err)
	}
	params.Cluster = cluster

	return nil
}
