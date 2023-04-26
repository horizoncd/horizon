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
	applicationgitrepo "github.com/horizoncd/horizon/pkg/application/gitrepo"
	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	"github.com/horizoncd/horizon/pkg/cd"
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
	CD                   cd.CD
	K8sUtil              cd.K8sUtil
	OutputGetter         output.Getter
	TektonFty            factory.Factory
	ClusterGitRepo       clustergitrepo.ClusterGitRepo
	GitGetter            code.GitGetter
	BuildSchema          *build.Schema
}
