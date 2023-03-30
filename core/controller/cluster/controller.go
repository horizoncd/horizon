package cluster

import (
	"context"

	"github.com/horizoncd/horizon/core/config"
	"github.com/horizoncd/horizon/core/controller/build"
	"github.com/horizoncd/horizon/lib/q"
	appgitrepo "github.com/horizoncd/horizon/pkg/application/gitrepo"
	appmanager "github.com/horizoncd/horizon/pkg/application/manager"
	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	"github.com/horizoncd/horizon/pkg/cluster/cd"
	"github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	registryfty "github.com/horizoncd/horizon/pkg/cluster/registry/factory"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	collectionmanager "github.com/horizoncd/horizon/pkg/collection/manager"
	"github.com/horizoncd/horizon/pkg/config/grafana"
	"github.com/horizoncd/horizon/pkg/config/template"
	"github.com/horizoncd/horizon/pkg/config/token"
	envmanager "github.com/horizoncd/horizon/pkg/environment/manager"
	"github.com/horizoncd/horizon/pkg/environment/service"
	environmentregionmapper "github.com/horizoncd/horizon/pkg/environmentregion/manager"
	eventmanager "github.com/horizoncd/horizon/pkg/event/manager"
	grafanaservice "github.com/horizoncd/horizon/pkg/grafana"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	groupsvc "github.com/horizoncd/horizon/pkg/group/service"
	"github.com/horizoncd/horizon/pkg/member"
	"github.com/horizoncd/horizon/pkg/param"
	prmanager "github.com/horizoncd/horizon/pkg/pipelinerun/manager"
	pipelinemanager "github.com/horizoncd/horizon/pkg/pipelinerun/pipeline/manager"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	"github.com/horizoncd/horizon/pkg/templaterelease/output"
	templateschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
	templateschematagmanager "github.com/horizoncd/horizon/pkg/templateschematag/manager"
	tokenservice "github.com/horizoncd/horizon/pkg/token/service"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usersvc "github.com/horizoncd/horizon/pkg/user/service"
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
	List(ctx context.Context, query *q.Query) ([]*ListClusterWithFullResponse, int, error)
	ListByApplication(ctx context.Context, query *q.Query) (int, []*ListClusterResponse, error)
	ListClusterWithExpiry(ctx context.Context, query *q.Query) ([]*ListClusterWithExpiryResponse, error)

	BuildDeploy(ctx context.Context, clusterID uint,
		request *BuildDeployRequest) (*BuildDeployResponse, error)
	Restart(ctx context.Context, clusterID uint) (*PipelinerunIDResponse, error)
	Deploy(ctx context.Context, clusterID uint, request *DeployRequest) (*PipelinerunIDResponse, error)
	Rollback(ctx context.Context, clusterID uint, request *RollbackRequest) (*PipelinerunIDResponse, error)

	FreeCluster(ctx context.Context, clusterID uint) error

	// InternalDeploy todo(zx): remove after InternalDeployV2 is stabilized
	InternalDeploy(ctx context.Context, clusterID uint,
		r *InternalDeployRequest) (_ *InternalDeployResponse, err error)

	Promote(ctx context.Context, clusterID uint) error
	Pause(ctx context.Context, clusterID uint) error
	Resume(ctx context.Context, clusterID uint) error
	Next(ctx context.Context, clusterID uint) error

	// Deprecated: GetClusterStatus
	GetClusterStatus(ctx context.Context, clusterID uint) (_ *GetClusterStatusResponse, err error)
	// Deprecated
	Online(ctx context.Context, clusterID uint, r *ExecRequest) (ExecResponse, error)
	// Deprecated
	Offline(ctx context.Context, clusterID uint, r *ExecRequest) (ExecResponse, error)
	Exec(ctx context.Context, clusterID uint, r *ExecRequest) (_ ExecResponse, err error)

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
	// InternalDeployV2 deploy only used by internal system
	InternalDeployV2(ctx context.Context, clusterID uint,
		r *InternalDeployRequestV2) (_ *InternalDeployResponseV2, err error)
	InternalGetClusterStatus(ctx context.Context, clusterID uint) (_ *GetClusterStatusResponse, err error)
	GetClusterStatusV2(ctx context.Context, clusterID uint) (_ *StatusResponseV2, err error)
	GetClusterBuildStatus(ctx context.Context, clusterID uint) (*BuildStatusResponse, error)
	GetResourceTree(ctx context.Context, clusterID uint) (*GetResourceTreeResponse, error)
	GetStep(ctx context.Context, clusterID uint) (resp *GetStepResponse, err error)
	// Deprecated: for internal usage, v1 to v2
	Upgrade(ctx context.Context, clusterID uint) error
	ToggleLikeStatus(ctx context.Context, clusterID uint, like *WhetherLike) (err error)
}

type controller struct {
	clusterMgr            clustermanager.Manager
	clusterGitRepo        gitrepo.ClusterGitRepo
	applicationGitRepo    appgitrepo.ApplicationGitRepo
	commitGetter          code.GitGetter
	cd                    cd.CD
	applicationMgr        appmanager.Manager
	autoFreeSvc           *service.AutoFreeSVC
	applicationSvc        applicationservice.Service
	templateReleaseMgr    trmanager.Manager
	templateSchemaGetter  templateschema.Getter
	outputGetter          output.Getter
	envMgr                envmanager.Manager
	envRegionMgr          environmentregionmapper.Manager
	regionMgr             regionmanager.Manager
	groupSvc              groupsvc.Service
	pipelinerunMgr        prmanager.Manager
	pipelineMgr           pipelinemanager.Manager
	tektonFty             factory.Factory
	registryFty           registryfty.RegistryGetter
	userManager           usermanager.Manager
	userSvc               usersvc.Service
	memberManager         member.Manager
	groupManager          groupmanager.Manager
	schemaTagManager      templateschematagmanager.Manager
	tagMgr                tagmanager.Manager
	grafanaService        grafanaservice.Service
	grafanaConfig         grafana.Config
	buildSchema           *build.Schema
	eventMgr              eventmanager.Manager
	tokenSvc              tokenservice.Service
	tokenConfig           token.Config
	templateUpgradeMapper template.UpgradeMapper
	collectionManager     collectionmanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(config *config.Config, param *param.Param) Controller {
	return &controller{
		clusterMgr:            param.ClusterMgr,
		clusterGitRepo:        param.ClusterGitRepo,
		applicationGitRepo:    param.ApplicationGitRepo,
		commitGetter:          param.GitGetter,
		cd:                    param.Cd,
		applicationMgr:        param.ApplicationManager,
		applicationSvc:        param.ApplicationSvc,
		templateReleaseMgr:    param.TemplateReleaseManager,
		templateSchemaGetter:  param.TemplateSchemaGetter,
		autoFreeSvc:           param.AutoFreeSvc,
		outputGetter:          param.OutputGetter,
		envMgr:                param.EnvMgr,
		envRegionMgr:          param.EnvRegionMgr,
		regionMgr:             param.RegionMgr,
		groupSvc:              param.GroupSvc,
		pipelinerunMgr:        param.PipelinerunMgr,
		pipelineMgr:           param.PipelineMgr,
		tektonFty:             param.TektonFty,
		registryFty:           registryfty.Fty,
		userManager:           param.UserManager,
		userSvc:               param.UserSvc,
		memberManager:         param.MemberManager,
		groupManager:          param.GroupManager,
		schemaTagManager:      param.ClusterSchemaTagMgr,
		tagMgr:                param.TagManager,
		grafanaService:        param.GrafanaService,
		grafanaConfig:         config.GrafanaConfig,
		buildSchema:           param.BuildSchema,
		eventMgr:              param.EventManager,
		tokenSvc:              param.TokenSvc,
		tokenConfig:           config.TokenConfig,
		templateUpgradeMapper: config.TemplateUpgradeMapper,
		collectionManager:     param.CollectionMgr,
	}
}
