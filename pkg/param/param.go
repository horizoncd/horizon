// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package param

import (
	"github.com/horizoncd/horizon/core/controller/build"
	"github.com/horizoncd/horizon/pkg/cd"
	"github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	"github.com/horizoncd/horizon/pkg/gitrepo"
	"github.com/horizoncd/horizon/pkg/grafana"
	"github.com/horizoncd/horizon/pkg/hook/hook"
	oauthmanager "github.com/horizoncd/horizon/pkg/manager"
	"github.com/horizoncd/horizon/pkg/oauth/scope"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/service"
	"github.com/horizoncd/horizon/pkg/templaterelease/output"
	templateschema "github.com/horizoncd/horizon/pkg/templaterelease/schema"
)

type Param struct {
	// manager
	*managerparam.Manager

	OauthManager oauthmanager.OAuthManager
	// service
	AutoFreeSvc    *service.AutoFreeSVC
	MemberService  service.MemberService
	ApplicationSvc service.ApplicationService
	ClusterSvc     service.ClusterService
	GroupSvc       service.GroupService
	UserSvc        service.UserService
	TokenSvc       service.TokenService
	RoleService    role.Service
	ScopeService   scope.Service
	GrafanaService grafana.Service

	// others
	Hook                 hook.Hook
	ApplicationGitRepo   gitrepo.ApplicationGitRepo
	TemplateSchemaGetter templateschema.Getter
	CD                   cd.CD
	K8sUtil              cd.K8sUtil
	OutputGetter         output.Getter
	TektonFty            factory.Factory
	ClusterGitRepo       gitrepo.ClusterGitRepo
	GitGetter            code.GitGetter
	BuildSchema          *build.Schema
}
