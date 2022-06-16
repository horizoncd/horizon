package managerparam

import (
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	applicationregionmanager "g.hz.netease.com/horizon/pkg/applicationregion/manager"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	environmentregionmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	harbormanager "g.hz.netease.com/horizon/pkg/harbor/manager"
	membermanager "g.hz.netease.com/horizon/pkg/member"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	pipelinemanager "g.hz.netease.com/horizon/pkg/pipelinerun/pipeline/manager"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	tagmanager "g.hz.netease.com/horizon/pkg/tag/manager"
	templatemanager "g.hz.netease.com/horizon/pkg/template/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	templateschematagmanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"gorm.io/gorm"
)

type Manager struct {
	UserManager              usermanager.Manager
	ApplicationManager       applicationmanager.Manager
	TemplateReleaseManager   trmanager.Manager
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
	PipelinerMgr             pipelinemanager.Manager
	EnvMgr                   envmanager.Manager
	GroupManager             groupmanager.Manager
	HarborManager            harbormanager.Manager
}

func InitManager(db *gorm.DB) *Manager {
	return &Manager{
		UserManager:              usermanager.New(db),
		ApplicationManager:       applicationmanager.New(db),
		TemplateReleaseManager:   trmanager.New(db),
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
		PipelinerMgr:             pipelinemanager.New(db),
		EnvMgr:                   envmanager.New(db),
		GroupManager:             groupmanager.New(db),
		HarborManager:            harbormanager.New(db),
	}
}
