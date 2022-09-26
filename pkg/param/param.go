package param

import (
	applicationgitrepo "g.hz.netease.com/horizon/pkg/application/gitrepo"
	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	clustergitrepo "g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clusterservice "g.hz.netease.com/horizon/pkg/cluster/service"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/grafana"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	memberservice "g.hz.netease.com/horizon/pkg/member/service"
	oauthmanager "g.hz.netease.com/horizon/pkg/oauth/manager"
	"g.hz.netease.com/horizon/pkg/oauth/scope"
	"g.hz.netease.com/horizon/pkg/param/managerparam"

	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/templaterelease/output"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	userservice "g.hz.netease.com/horizon/pkg/user/service"
)

type Param struct {
	// manager
	*managerparam.Manager

	OauthManager oauthmanager.Manager
	// service
	MemberService  memberservice.Service
	ApplicationSvc applicationservice.Service
	ClusterSvc     clusterservice.Service
	GroupSvc       groupsvc.Service
	UserSvc        userservice.Service
	RoleService    role.Service
	ScopeService   scope.Service
	GrafanaService grafana.Service

	// others
	Hook                 hook.Hook
	ApplicationGitRepo   applicationgitrepo.ApplicationGitRepo2
	TemplateSchemaGetter templateschema.Getter
	Cd                   cd.CD
	OutputGetter         output.Getter
	TektonFty            factory.Factory
	ClusterGitRepo       clustergitrepo.ClusterGitRepo
	GitGetter            code.GitGetter
}
