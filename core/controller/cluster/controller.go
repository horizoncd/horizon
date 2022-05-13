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
	"g.hz.netease.com/horizon/pkg/config/grafana"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	environmentregionmapper "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/member"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/templaterelease/output"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	tagmanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usersvc "g.hz.netease.com/horizon/pkg/user/service"
)

type Controller interface {
	GetCluster(ctx context.Context, clusterID uint) (*GetClusterResponse, error)
	GetClusterOutput(ctx context.Context, clusterID uint) (interface{}, error)
	ListCluster(ctx context.Context, applicationID uint, environments []string,
		filter string, query *q.Query) (int, []*ListClusterResponse, error)
	ListClusterByNameFuzzily(ctx context.Context, environment,
		filter string, query *q.Query) (int, []*ListClusterWithFullResponse, error)
	ListUserClusterByNameFuzzily(ctx context.Context, environment,
		filter string, query *q.Query) (int, []*ListClusterWithFullResponse, error)
	CreateCluster(ctx context.Context, applicationID uint, environment, region string,
		request *CreateClusterRequest) (*GetClusterResponse, error)
	UpdateCluster(ctx context.Context, clusterID uint,
		request *UpdateClusterRequest) (*GetClusterResponse, error)
	DeleteCluster(ctx context.Context, clusterID uint) error
	GetClusterByName(ctx context.Context,
		clusterName string) (*GetClusterByNameResponse, error)

	BuildDeploy(ctx context.Context, clusterID uint,
		request *BuildDeployRequest) (*BuildDeployResponse, error)
	GetDiff(ctx context.Context, clusterID uint, codeBranch string) (*GetDiffResponse, error)
	GetClusterStatus(ctx context.Context, clusterID uint) (_ *GetClusterStatusResponse, err error)
	Restart(ctx context.Context, clusterID uint) (*PipelinerunIDResponse, error)
	Deploy(ctx context.Context, clusterID uint, request *DeployRequest) (*PipelinerunIDResponse, error)
	Rollback(ctx context.Context, clusterID uint, request *RollbackRequest) (*PipelinerunIDResponse, error)
	Next(ctx context.Context, clusterID uint) error
	GetContainerLog(ctx context.Context, clusterID uint, podName, containerName string, tailLines int) (
		<-chan string, error)
	Online(ctx context.Context, clusterID uint, r *ExecRequest) (ExecResponse, error)
	Offline(ctx context.Context, clusterID uint, r *ExecRequest) (ExecResponse, error)
	GetDashboard(ctx context.Context, clusterID uint) (*GetDashboardResponse, error)
	GetClusterPods(ctx context.Context, clusterID uint, start, end int64) (*GetClusterPodsResponse, error)
	FreeCluster(ctx context.Context, clusterID uint) error
	GetPodEvents(ctx context.Context, clusterID uint, podName string) (interface{}, error)
	Promote(ctx context.Context, clusterID uint) error
	Pause(ctx context.Context, clusterID uint) error
	Resume(ctx context.Context, clusterID uint) error
	GetContainers(ctx context.Context, clusterID uint, podName string) (interface{}, error)
	// InternalDeploy deploy only used by internal system
	InternalDeploy(ctx context.Context, clusterID uint,
		r *InternalDeployRequest) (_ *InternalDeployResponse, err error)
}

type controller struct {
	clusterMgr           clustermanager.Manager
	clusterGitRepo       gitrepo.ClusterGitRepo
	applicationGitRepo   appgitrepo.ApplicationGitRepo
	commitGetter         code.GitGetter
	cd                   cd.CD
	applicationMgr       appmanager.Manager
	applicationSvc       applicationservice.Service
	templateReleaseMgr   trmanager.Manager
	templateSchemaGetter templateschema.Getter
	outputGetter         output.Getter
	envMgr               envmanager.Manager
	envRegionMgr         environmentregionmapper.EnvironmentRegionManager
	regionMgr            regionmanager.Manager
	groupSvc             groupsvc.Service
	hook                 hook.Hook
	pipelinerunMgr       prmanager.Manager
	tektonFty            factory.Factory
	registryFty          registryfty.Factory
	grafanaMapper        grafana.Mapper
	userManager          usermanager.Manager
	userSvc              usersvc.Service
	memberManager        member.Manager
	groupManager         groupmanager.Manager
	tagManager           tagmanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(clusterGitRepo gitrepo.ClusterGitRepo, applicationGitRepo appgitrepo.ApplicationGitRepo,
	commitGetter code.GitGetter, cd cd.CD, tektonFty factory.Factory,
	templateSchemaGetter templateschema.Getter, outputGetter output.Getter,
	hook hook.Hook, grafanaMapper grafana.Mapper, tagManager tagmanager.Manager) Controller {
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
		outputGetter:         outputGetter,
		envMgr:               envmanager.Mgr,
		envRegionMgr:         environmentregionmapper.Mgr,
		regionMgr:            regionmanager.Mgr,
		groupSvc:             groupsvc.Svc,
		pipelinerunMgr:       prmanager.Mgr,
		tektonFty:            tektonFty,
		registryFty:          registryfty.Fty,
		hook:                 hook,
		grafanaMapper:        grafanaMapper,
		userManager:          usermanager.Mgr,
		userSvc:              usersvc.Svc,
		memberManager:        member.Mgr,
		groupManager:         groupmanager.Mgr,
		tagManager:           tagManager,
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
