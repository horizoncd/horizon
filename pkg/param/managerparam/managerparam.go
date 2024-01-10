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

package managerparam

import (
	"gorm.io/gorm"

	collectionmanager "github.com/horizoncd/horizon/pkg/collection/manager"

	accesstokenmanager "github.com/horizoncd/horizon/pkg/accesstoken/manager"
	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	applicationregionmanager "github.com/horizoncd/horizon/pkg/applicationregion/manager"
	badgemanager "github.com/horizoncd/horizon/pkg/badge/manager"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	envmanager "github.com/horizoncd/horizon/pkg/environment/manager"
	environmentregionmanager "github.com/horizoncd/horizon/pkg/environmentregion/manager"
	eventManager "github.com/horizoncd/horizon/pkg/event/manager"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	idpmanager "github.com/horizoncd/horizon/pkg/idp/manager"
	membermanager "github.com/horizoncd/horizon/pkg/member"
	prmanager "github.com/horizoncd/horizon/pkg/pr/manager"
	pipelinemanager "github.com/horizoncd/horizon/pkg/pr/pipeline/manager"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	registrymanager "github.com/horizoncd/horizon/pkg/registry/manager"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	templatemanager "github.com/horizoncd/horizon/pkg/template/manager"
	trmanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	templateschematagmanager "github.com/horizoncd/horizon/pkg/templateschematag/manager"
	trtmanager "github.com/horizoncd/horizon/pkg/templateschematag/manager"
	tokenmanager "github.com/horizoncd/horizon/pkg/token/manager"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	linkmanager "github.com/horizoncd/horizon/pkg/userlink/manager"
	webhookManager "github.com/horizoncd/horizon/pkg/webhook/manager"
)

type Manager struct {
	UserMgr              usermanager.Manager
	UserLinksMgr         linkmanager.Manager
	ApplicationMgr       applicationmanager.Manager
	TemplateReleaseMgr   trmanager.Manager
	TemplateSchemaTagMgr trtmanager.Manager
	CollectionMgr        collectionmanager.Manager
	ClusterMgr           clustermanager.Manager
	MemberMgr            membermanager.Manager
	ClusterSchemaTagMgr  templateschematagmanager.Manager
	ApplicationRegionMgr applicationregionmanager.Manager
	EnvironmentRegionMgr environmentregionmanager.Manager
	TagMgr               tagmanager.Manager
	TemplateMgr          templatemanager.Manager
	EnvRegionMgr         environmentregionmanager.Manager
	RegionMgr            regionmanager.Manager
	PRMgr                *prmanager.PRManager
	PipelineMgr          pipelinemanager.Manager
	EnvMgr               envmanager.Manager
	GroupMgr             groupmanager.Manager
	RegistryMgr          registrymanager.Manager
	IdpMgr               idpmanager.Manager
	AccessTokenMgr       accesstokenmanager.Manager
	WebhookMgr           webhookManager.Manager
	EventMgr             eventManager.Manager
	TokenMgr             tokenmanager.Manager
	BadgeMgr             badgemanager.Manager
}

func InitManager(db *gorm.DB) *Manager {
	return &Manager{
		UserMgr:              usermanager.New(db),
		UserLinksMgr:         linkmanager.New(db),
		ApplicationMgr:       applicationmanager.New(db),
		TemplateReleaseMgr:   trmanager.New(db),
		TemplateSchemaTagMgr: trtmanager.New(db),
		ClusterMgr:           clustermanager.New(db),
		CollectionMgr:        collectionmanager.New(db),
		MemberMgr:            membermanager.New(db),
		ClusterSchemaTagMgr:  templateschematagmanager.New(db),
		ApplicationRegionMgr: applicationregionmanager.New(db),
		EnvironmentRegionMgr: environmentregionmanager.New(db),
		TagMgr:               tagmanager.New(db),
		TemplateMgr:          templatemanager.New(db),
		EnvRegionMgr:         environmentregionmanager.New(db),
		RegionMgr:            regionmanager.New(db),
		PRMgr:                prmanager.NewPRManager(db),
		PipelineMgr:          pipelinemanager.New(db),
		EnvMgr:               envmanager.New(db),
		GroupMgr:             groupmanager.New(db),
		RegistryMgr:          registrymanager.New(db),
		IdpMgr:               idpmanager.NewManager(db),
		AccessTokenMgr:       accesstokenmanager.New(db),
		WebhookMgr:           webhookManager.New(db),
		EventMgr:             eventManager.New(db),
		TokenMgr:             tokenmanager.New(db),
		BadgeMgr:             badgemanager.New(db),
	}
}
