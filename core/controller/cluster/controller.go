package cluster

import (
	"context"

	"g.hz.netease.com/horizon/core/config"
	"g.hz.netease.com/horizon/core/controller/build"
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
	grafanaservice "g.hz.netease.com/horizon/pkg/grafana"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/member"
	"g.hz.netease.com/horizon/pkg/param"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	pipelinemanager "g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/manager"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	tagmanager "g.hz.netease.com/horizon/pkg/tag/manager"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/templaterelease/output"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	templateschematagmanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usersvc "g.hz.netease.com/horizon/pkg/user/service"
)

type Controller interface {
	CreateCluster(ctx context.Context, applicationID uint, environment, region string,
		request *CreateClusterRequest, mergePatch bool) (*GetClusterResponse, error)
	UpdateCluster(ctx context.Context, clusterID uint,
		request *UpdateClusterRequest, mergePatch bool) (*GetClusterResponse, error)
	DeleteCluster(ctx context.Context, clusterID uint, hard bool) error

	GetCluster(ctx context.Context, clusterID uint) (*GetClusterResponse, error)
	GetClusterByName(ctx context.Context,
		clusterName string) (*GetClusterByNameResponse, error)
	GetClusterOutput(ctx context.Context, clusterID uint) (interface{}, error)
	ListCluster(ctx context.Context, applicationID uint, environments []string,
		filter string, query *q.Query, ts []tagmodels.TagSelector) (int, []*ListClusterResponse, error)
	ListClusterByNameFuzzily(ctx context.Context, environment,
		filter string, query *q.Query) (int, []*ListClusterWithFullResponse, error)
	ListUserClusterByNameFuzzily(ctx context.Context, environment,
		filter string, query *q.Query) (int, []*ListClusterWithFullResponse, error)
	ListClusterWithExpiry(ctx context.Context, query *q.Query) ([]*ListClusterWithExpiryResponse, error)

	BuildDeploy(ctx context.Context, clusterID uint,
		request *BuildDeployRequest) (*BuildDeployResponse, error)
	Restart(ctx context.Context, clusterID uint) (*PipelinerunIDResponse, error)
	Deploy(ctx context.Context, clusterID uint, request *DeployRequest) (*PipelinerunIDResponse, error)
	Rollback(ctx context.Context, clusterID uint, request *RollbackRequest) (*PipelinerunIDResponse, error)

	FreeCluster(ctx context.Context, clusterID uint) error
	// InternalDeploy deploy only used by internal system
	InternalDeploy(ctx context.Context, clusterID uint,
		r *InternalDeployRequest) (_ *InternalDeployResponse, err error)

	Promote(ctx context.Context, clusterID uint) error
	Pause(ctx context.Context, clusterID uint) error
	Resume(ctx context.Context, clusterID uint) error
	Next(ctx context.Context, clusterID uint) error

	GetClusterStatus(ctx context.Context, clusterID uint) (_ *GetClusterStatusResponse, err error)
	Online(ctx context.Context, clusterID uint, r *ExecRequest) (ExecResponse, error)
	Offline(ctx context.Context, clusterID uint, r *ExecRequest) (ExecResponse, error)

	GetDiff(ctx context.Context, clusterID uint, refType, ref string) (*GetDiffResponse, error)
	GetContainerLog(ctx context.Context, clusterID uint, podName, containerName string, tailLines int) (
		<-chan string, error)

	DeleteClusterPods(ctx context.Context, clusterID uint, podName []string) (BatchResponse, error)
	GetClusterPod(ctx context.Context, clusterID uint, podName string) (
		*GetClusterPodResponse, error)

	GetPodEvents(ctx context.Context, clusterID uint, podName string) (interface{}, error)
	GetContainers(ctx context.Context, clusterID uint, podName string) (interface{}, error)
	GetGrafanaDashBoard(c context.Context, clusterID uint) (*GetGrafanaDashboardsResponse, error)

	CreateClusterV2(ctx context.Context, applicationID uint, environment,
		region string, r *CreateClusterRequestV2, mergePatch bool) (*CreateClusterResponseV2, error)
	GetClusterV2(ctx context.Context, clusterID uint) (*GetClusterResponseV2, error)
	UpdateClusterV2(ctx context.Context, clusterID uint, r *UpdateClusterRequestV2, mergePatch bool) error
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
	envRegionMgr         environmentregionmapper.Manager
	regionMgr            regionmanager.Manager
	groupSvc             groupsvc.Service
	hook                 hook.Hook
	pipelinerunMgr       prmanager.Manager
	pipelineMgr          pipelinemanager.Manager
	tektonFty            factory.Factory
	registryFty          registryfty.Factory
	userManager          usermanager.Manager
	userSvc              usersvc.Service
	memberManager        member.Manager
	groupManager         groupmanager.Manager
	schemaTagManager     templateschematagmanager.Manager
	tagMgr               tagmanager.Manager
	grafanaService       grafanaservice.Service
	grafanaConfig        grafana.Config
	buildSchema          *build.Schema
}

var _ Controller = (*controller)(nil)

func NewController(config *config.Config, param *param.Param) Controller {
	return &controller{
		clusterMgr:           param.ClusterMgr,
		clusterGitRepo:       param.ClusterGitRepo,
		applicationGitRepo:   param.ApplicationGitRepo,
		commitGetter:         param.GitGetter,
		cd:                   param.Cd,
		applicationMgr:       param.ApplicationManager,
		applicationSvc:       param.ApplicationSvc,
		templateReleaseMgr:   param.TemplateReleaseManager,
		templateSchemaGetter: param.TemplateSchemaGetter,
		outputGetter:         param.OutputGetter,
		envMgr:               param.EnvMgr,
		envRegionMgr:         param.EnvRegionMgr,
		regionMgr:            param.RegionMgr,
		groupSvc:             param.GroupSvc,
		pipelinerunMgr:       param.PipelinerunMgr,
		pipelineMgr:          param.PipelineMgr,
		tektonFty:            param.TektonFty,
		registryFty:          registryfty.Fty,
		hook:                 param.Hook,
		userManager:          param.UserManager,
		userSvc:              param.UserSvc,
		memberManager:        param.MemberManager,
		groupManager:         param.GroupManager,
		schemaTagManager:     param.ClusterSchemaTagMgr,
		tagMgr:               param.TagManager,
		grafanaService:       param.GrafanaService,
		grafanaConfig:        config.GrafanaConfig,
		buildSchema:          param.BuildSchema,
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
