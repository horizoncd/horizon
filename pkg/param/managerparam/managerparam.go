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
	"github.com/horizoncd/horizon/pkg/manager"
	"gorm.io/gorm"
)

type Manager struct {
	UserManager              manager.UserManager
	UserLinksManager         manager.UserLinkManager
	ApplicationManager       manager.ApplicationManager
	TemplateReleaseManager   manager.TemplateReleaseManager
	TemplateSchemaTagManager manager.TemplateSchemaTagManager
	CollectionMgr            manager.CollectionManager
	ClusterMgr               manager.ClusterManager
	MemberManager            manager.MemberManager
	ClusterSchemaTagMgr      manager.TemplateSchemaTagManager
	ApplicationRegionManager manager.ApplicationRegionManager
	EnvironmentRegionMgr     manager.EnvironmentRegionManager
	TagManager               manager.TagManager
	TemplateMgr              manager.TemplateManager
	EnvRegionMgr             manager.EnvironmentRegionManager
	RegionMgr                manager.RegionManager
	PipelinerunMgr           manager.PipelineRunManager
	PipelineMgr              manager.PipelineManager
	EnvMgr                   manager.EnvironmentManager
	GroupManager             manager.GroupManager
	RegistryManager          manager.RegistryManager
	IdpManager               manager.IDProviderManager
	AccessTokenManager       manager.AccessTokenManager
	WebhookManager           manager.WebhookManager
	EventManager             manager.EventManager
	TokenManager             manager.TokenManager
}

func InitManager(db *gorm.DB) *Manager {
	return &Manager{
		UserManager:              manager.NewUserManager(db),
		UserLinksManager:         manager.NewUserLinkManager(db),
		ApplicationManager:       manager.NewApplicationManager(db),
		TemplateReleaseManager:   manager.NewTemplateReleaseManager(db),
		TemplateSchemaTagManager: manager.NewTemplateSchemaTagManager(db),
		ClusterMgr:               manager.NewClusterManager(db),
		CollectionMgr:            manager.NewCollectionManager(db),
		MemberManager:            manager.NewMemberManager(db),
		ClusterSchemaTagMgr:      manager.NewTemplateSchemaTagManager(db),
		ApplicationRegionManager: manager.NewApplicationRegionManager(db),
		EnvironmentRegionMgr:     manager.NewEnvironmentRegionManager(db),
		TagManager:               manager.NewTagManager(db),
		TemplateMgr:              manager.NewTemplateManager(db),
		EnvRegionMgr:             manager.NewEnvironmentRegionManager(db),
		RegionMgr:                manager.NewRegionManager(db),
		PipelinerunMgr:           manager.NewPipelineRunManager(db),
		PipelineMgr:              manager.NewPipelineManager(db),
		EnvMgr:                   manager.NewEnvironmentManager(db),
		GroupManager:             manager.NewGroupManager(db),
		RegistryManager:          manager.NewRegistryManager(db),
		IdpManager:               manager.NewIDProviderManager(db),
		AccessTokenManager:       manager.NewAccessTokenManager(db),
		WebhookManager:           manager.NewWebhookManager(db),
		EventManager:             manager.NewEventManager(db),
		TokenManager:             manager.NewTokenManager(db),
	}
}
