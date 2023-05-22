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

	"github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/util/sets"
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

type Tags []*Tag

func (t Tags) IntoTagsBasic() TagsBasic {
	tags := make(TagsBasic, 0, len(t))
	for _, tag := range t {
		tags = append(tags, &TagBasic{
			Key:   tag.Key,
			Value: tag.Value,
		})
	}
	return tags
}

func (t Tags) Eq(rhs Tags) bool {
	if len(t) != len(rhs) {
		return false
	}
	index := make(map[TagBasic]struct{})
	for _, tag := range t {
		index[TagBasic{
			Key:   tag.Key,
			Value: tag.Value,
		}] = struct{}{}
	}
	for _, tag := range rhs {
		if _, ok := index[TagBasic{
			Key:   tag.Key,
			Value: tag.Value,
		}]; !ok {
			return false
		}
	}
	return true
}

type TagsBasic []*TagBasic

func (t TagsBasic) Eq(rhs Tags) bool {
	if len(t) != len(rhs) {
		return false
	}
	index := make(map[TagBasic]struct{})
	for _, tag := range t {
		index[*tag] = struct{}{}
	}
	for _, tag := range rhs {
		if _, ok := index[TagBasic{
			Key:   tag.Key,
			Value: tag.Value,
		}]; !ok {
			return false
		}
	}
	return true
}

func (t TagsBasic) IntoTags(resourceType models.ResourceType, resourceID uint) []*Tag {
	tags := make([]*Tag, 0, len(t))
	for _, tag := range t {
		tags = append(tags, &Tag{
			ResourceType: string(resourceType),
			ResourceID:   resourceID,
			Key:          tag.Key,
			Value:        tag.Value,
		})
	}
	return tags
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

type Metatag struct {
	TagKey      string    `json:"tagKey"`
	TagValue    string    `json:"tagValue"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}
