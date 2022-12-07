package managerparam

import (
	"gorm.io/gorm"

	accesstokenmanager "github.com/horizoncd/horizon/pkg/accesstoken/manager"
	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	applicationregionmanager "github.com/horizoncd/horizon/pkg/applicationregion/manager"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	envmanager "github.com/horizoncd/horizon/pkg/environment/manager"
	environmentregionmanager "github.com/horizoncd/horizon/pkg/environmentregion/manager"
	eventManager "github.com/horizoncd/horizon/pkg/event/manager"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	idpmanager "github.com/horizoncd/horizon/pkg/idp/manager"
	membermanager "github.com/horizoncd/horizon/pkg/member"
	prmanager "github.com/horizoncd/horizon/pkg/pipelinerun/manager"
	pipelinemanager "github.com/horizoncd/horizon/pkg/pipelinerun/pipeline/manager"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	registrymanager "github.com/horizoncd/horizon/pkg/registry/manager"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	templatemanager "github.com/horizoncd/horizon/pkg/template/manager"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	templateschematagmanager "github.com/horizoncd/horizon/pkg/templateschematag/manager"
	trtmanager "github.com/horizoncd/horizon/pkg/templateschematag/manager"
	tokenmanager "g.hz.netease.com/horizon/pkg/token/manager"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	linkmanager "github.com/horizoncd/horizon/pkg/userlink/manager"
	webhookManager "github.com/horizoncd/horizon/pkg/webhook/manager"
)

type Manager struct {
	UserManager              usermanager.Manager
	UserLinksManager         linkmanager.Manager
	ApplicationManager       applicationmanager.Manager
	TemplateReleaseManager   trmanager.Manager
	TemplateSchemaTagManager trtmanager.Manager
	ClusterMgr               clustermanager.Manager
	MemberManager            membermanager.Manager
	ClusterSchemaTagMgr      templateschematagmanager.Manager
	ApplicationRegionManager applicationregionmanager.Manager
	EnvironmentRegionMgr     environmentregionmanager.Manager
	TagManager               tagmanager.Manager
	TemplateMgr              templatemanager.Manager
	EnvRegionMgr             environmentregionmanager.Manager
	RegionMgr                regionmanager.Manager
	PipelinerunMgr           prmanager.Manager
	PipelineMgr              pipelinemanager.Manager
	EnvMgr                   envmanager.Manager
	GroupManager             groupmanager.Manager
	RegistryManager          registrymanager.Manager
	IdpManager               idpmanager.Manager
	AccessTokenManager       accesstokenmanager.Manager
	WebhookManager           webhookManager.Manager
	EventManager             eventManager.Manager
	TokenManager             tokenmanager.Manager
}

func InitManager(db *gorm.DB) *Manager {
	return &Manager{
		UserManager:              usermanager.New(db),
		UserLinksManager:         linkmanager.New(db),
		ApplicationManager:       applicationmanager.New(db),
		TemplateReleaseManager:   trmanager.New(db),
		TemplateSchemaTagManager: trtmanager.New(db),
		ClusterMgr:               clustermanager.New(db),
		MemberManager:            membermanager.New(db),
		ClusterSchemaTagMgr:      templateschematagmanager.New(db),
		ApplicationRegionManager: applicationregionmanager.New(db),
		EnvironmentRegionMgr:     environmentregionmanager.New(db),
		TagManager:               tagmanager.New(db),
		TemplateMgr:              templatemanager.New(db),
		EnvRegionMgr:             environmentregionmanager.New(db),
		RegionMgr:                regionmanager.New(db),
		PipelinerunMgr:           prmanager.New(db),
		PipelineMgr:              pipelinemanager.New(db),
		EnvMgr:                   envmanager.New(db),
		GroupManager:             groupmanager.New(db),
		RegistryManager:          registrymanager.New(db),
		IdpManager:               idpmanager.NewManager(db),
		AccessTokenManager:       accesstokenmanager.New(db),
		WebhookManager:           webhookManager.New(db),
		EventManager:             eventManager.New(db),
		TokenManager:             tokenmanager.New(db),
	}
}
