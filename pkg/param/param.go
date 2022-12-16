package param

import (
	applicationgitrepo "github.com/horizoncd/horizon/pkg/application/gitrepo"
	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	"github.com/horizoncd/horizon/pkg/cluster/cd"
	"github.com/horizoncd/horizon/pkg/cluster/code"
	clustergitrepo "github.com/horizoncd/horizon/pkg/cluster/gitrepo"
	clusterservice "github.com/horizoncd/horizon/pkg/cluster/service"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	"github.com/horizoncd/horizon/pkg/environment/service"
	"github.com/horizoncd/horizon/pkg/grafana"
	groupsvc "github.com/horizoncd/horizon/pkg/group/service"
	"github.com/horizoncd/horizon/pkg/hook/hook"
	memberservice "github.com/horizoncd/horizon/pkg/member/service"
	oauthmanager "github.com/horizoncd/horizon/pkg/oauth/manager"
	"github.com/horizoncd/horizon/pkg/oauth/scope"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	tokenservice "github.com/horizoncd/horizon/pkg/token/service"

	"github.com/horizoncd/horizon/core/controller/build"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/templaterelease/output"
	templateschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
	userservice "github.com/horizoncd/horizon/pkg/user/service"
)

type Param struct {
	// manager
	*managerparam.Manager

	OauthManager oauthmanager.Manager
	// service
	AutoFreeSvc    *service.AutoFreeSVC
	MemberService  memberservice.Service
	ApplicationSvc applicationservice.Service
	ClusterSvc     clusterservice.Service
	GroupSvc       groupsvc.Service
	UserSvc        userservice.Service
	TokenSvc       tokenservice.Service
	RoleService    role.Service
	ScopeService   scope.Service
	GrafanaService grafana.Service

	// others
	Hook                 hook.Hook
	ApplicationGitRepo   applicationgitrepo.ApplicationGitRepo
	TemplateSchemaGetter templateschema.Getter
	Cd                   cd.CD
	OutputGetter         output.Getter
	TektonFty            factory.Factory
	ClusterGitRepo       clustergitrepo.ClusterGitRepo
	GitGetter            code.GitGetter
	BuildSchema          *build.Schema
}
