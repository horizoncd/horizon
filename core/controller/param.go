package controller

import (
	applicationgitrepo "g.hz.netease.com/horizon/pkg/application/gitrepo"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	applicationregionmanager "g.hz.netease.com/horizon/pkg/applicationregion/manager"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustergitrepo "g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/cluster/kubeclient"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	registryfty "g.hz.netease.com/horizon/pkg/cluster/registry/factory"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/config/grafana"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	environmentregionmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	environmentregionmapper "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/member"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	oauthmanager "g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/scope"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	tagmanager "g.hz.netease.com/horizon/pkg/tag/manager"
	tmanager "g.hz.netease.com/horizon/pkg/template/manager"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/templaterelease/output"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	templateschematagmanager "g.hz.netease.com/horizon/pkg/templateschematag/manager"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usersvc "g.hz.netease.com/horizon/pkg/user/service"
)

type Param struct {
	MemberService            memberservice.Service
	UserManager              usermanager.Manager
	ApplicationGitRepo       applicationgitrepo.ApplicationGitRepo
	TemplateSchemaGetter     templateschema.Getter
	ApplicationManager       applicationmanager.Manager
	ApplicationSvc           applicationservice.Service
	GroupMgr                 groupmanager.Manager
	GroupSvc                 groupsvc.Service
	TemplateReleaseManager   trmanager.Manager
	ClusterMgr               clustermanager.Manager
	Hook                     hook.Hook
	UserSvc                  usersvc.Service
	MemberManager            member.Manager
	EnvMgr                   envmanager.Manager
	CommitGetter             code.GitGetter
	Cd                       cd.CD
	OutputGetter             output.Getter
	EnvRegionMgr             environmentregionmapper.Manager
	RegionMgr                regionmanager.Manager
	PipelinerunMgr           prmanager.Manager
	TektonFty                factory.Factory
	RegistryFty              registryfty.Factory
	GrafanaMapper            grafana.Mapper
	GroupManager             groupmanager.Manager
	TagManager               tagmanager.Manager
	TemplateMgr              tmanager.Manager
	RoleService              role.Service
	KubeClientFty            kubeclient.Factory
	ClusterGitRepo           clustergitrepo.ClusterGitRepo
	GrafanaSLO               grafana.SLO
	GitGetter                code.GitGetter
	ClusterSchemaTagMgr      templateschematagmanager.Manager
	ApplicationRegionManager applicationregionmanager.Manager
	EnvironmentRegionMgr     environmentregionmanager.Manager
	OauthManager             oauthmanager.Manager
	ScopeService             scope.Service
}
