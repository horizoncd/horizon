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

package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/horizoncd/horizon/pkg/tag/models"
)

type Cluster struct {
	global.Model `json:"-"`

	ApplicationID   uint    `json:"application_id"`
	Name            string  `json:"name"`
	EnvironmentName string  `json:"environment_name"`
	RegionName      string  `json:"region_name"`
	Description     string  `json:"description"`
	Type            *string `gorm:"->"`
	GitURL          string  `json:"git_url"`
	GitSubfolder    string  `json:"git_subfolder"`
	GitRef          string  `json:"git_ref"`
	GitRefType      string  `json:"git_ref_type"`
	Image           string  `json:"image"`
	Template        string  `json:"template"`
	TemplateRelease string  `json:"template_release"`
	Status          string  `json:"status"`
	CreatedBy       uint    `json:"-"`
	UpdatedBy       uint    `json:"-"`
	ExpireSeconds   uint    `json:"-"`
}

func (c *Cluster) Copy() *Cluster {
	newCluster := &Cluster{
		Model: global.Model{
			ID: c.ID,
		},
		ApplicationID:   c.ApplicationID,
		Name:            c.Name,
		EnvironmentName: c.EnvironmentName,
		RegionName:      c.RegionName,
		Description:     c.Description,
		GitURL:          c.GitURL,
		GitSubfolder:    c.GitSubfolder,
		GitRef:          c.GitRef,
		GitRefType:      c.GitRefType,
		Image:           c.Image,
		Template:        c.Template,
		TemplateRelease: c.TemplateRelease,
		Status:          c.Status,
		ExpireSeconds:   c.ExpireSeconds,
	}
	return newCluster
}

type ClusterWithTags struct {
	*Cluster
	Tags []*models.Tag
}

type ClusterWithRegion struct {
	*Cluster

	RegionDisplayName string
}
