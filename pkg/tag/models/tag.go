/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package models

import (
	"time"

	"g.hz.netease.com/horizon/pkg/util/sets"
)

const (
	TypeGroup       = "groups"
	TypeApplication = "applications"
	TypeCluster     = "clusters"
	TypeRegion      = "regions"
)

type Tag struct {
	ID           uint   `gorm:"primarykey"`
	ResourceID   uint   `gorm:"uniqueIndex:idx_resource_key"`
	ResourceType string `gorm:"uniqueIndex:idx_resource_key"`
	Key          string `gorm:"uniqueIndex:idx_resource_key;column:tag_key"`
	Value        string `gorm:"column:tag_value"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CreatedBy    uint
	UpdatedBy    uint
}

type TagBasic struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type TagSelector struct {
	Key      string      `json:"key"`
	Values   sets.String `json:"values"`
	Operator string      `json:"operator"`
}

const (
	DoesNotExist string = "!"
	Equals       string = "="
	In           string = "in"
	NotEquals    string = "!="
	NotIn        string = "notin"
	Exists       string = "exists"
)
