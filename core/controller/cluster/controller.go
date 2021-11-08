package cluster

import (
	"context"

	"g.hz.netease.com/horizon/lib/q"
	appgitrepo "g.hz.netease.com/horizon/pkg/application/gitrepo"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	registryfty "g.hz.netease.com/horizon/pkg/cluster/registry/factory"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
)

type Controller interface {
	GetCluster(ctx context.Context, clusterID uint) (*GetClusterResponse, error)
	ListCluster(ctx context.Context, applicationID uint, environment,
		filter string, query *q.Query) (int, []*ListClusterResponse, error)
	ListClusterByNameFuzzily(ctx context.Context, environment,
		filter string, query *q.Query) (int, []*ListClusterWithFullResponse, error)
	CreateCluster(ctx context.Context, applicationID uint, environment, region string,
		request *CreateClusterRequest) (*GetClusterResponse, error)
	UpdateCluster(ctx context.Context, clusterID uint,
		request *UpdateClusterRequest) (*GetClusterResponse, error)
	GetClusterByName(ctx context.Context,
		clusterName string) (*GetClusterByNameResponse, error)

	BuildDeploy(ctx context.Context, clusterID uint,
		request *BuildDeployRequest) (*BuildDeployResponse, error)
	GetClusterStatus(ctx context.Context, clusterID uint) (_ *GetClusterStatusResponse, err error)

	// InternalDeploy deploy only used by internal system
	InternalDeploy(ctx context.Context, clusterID uint,
		r *InternalDeployRequest) (_ *InternalDeployResponse, err error)
}

type controller struct {
	clusterMgr           clustermanager.Manager
	clusterGitRepo       gitrepo.ClusterGitRepo
	applicationGitRepo   appgitrepo.ApplicationGitRepo
	commitGetter         code.CommitGetter
	cd                   cd.CD
	applicationMgr       appmanager.Manager
	applicationSvc       applicationservice.Service
	templateReleaseMgr   trmanager.Manager
	templateSchemaGetter templateschema.Getter
	envMgr               envmanager.Manager
	regionMgr            regionmanager.Manager
	groupSvc             groupsvc.Service
	hook                 hook.Hook
	pipelinerunMgr       prmanager.Manager
	tektonFty            factory.Factory
	registryFty          registryfty.Factory
}

var _ Controller = (*controller)(nil)

func NewController(clusterGitRepo gitrepo.ClusterGitRepo, applicationGitRepo appgitrepo.ApplicationGitRepo,
	commitGetter code.CommitGetter, cd cd.CD, tektonFty factory.Factory,
	templateSchemaGetter templateschema.Getter, hook hook.Hook) Controller {
	return &controller{
		clusterMgr:           clustermanager.Mgr,
		clusterGitRepo:       clusterGitRepo,
		applicationGitRepo:   applicationGitRepo,
		commitGetter:         commitGetter,
		cd:                   cd,
		applicationMgr:       appmanager.Mgr,
		applicationSvc:       applicationservice.Svc,
		templateReleaseMgr:   trmanager.Mgr,
		templateSchemaGetter: templateSchemaGetter,
		envMgr:               envmanager.Mgr,
		regionMgr:            regionmanager.Mgr,
		groupSvc:             groupsvc.Svc,
		pipelinerunMgr:       prmanager.Mgr,
		tektonFty:            tektonFty,
		registryFty:          registryfty.Fty,
		hook:                 hook,
	}
}

func (c *controller) postHook(ctx context.Context, eventType hook.EventType, content interface{}) {
	if c.hook != nil {
		event := hook.Event{
			EventType: eventType,
			Event:     content,
		}
		go c.hook.Push(ctx, event)
	}
}
