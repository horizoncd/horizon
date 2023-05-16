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

package tag

import (
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
)

type ListResponse struct {
	Tags []*Tag `json:"tags"`
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UpdateRequest struct {
	Tags []*Tag `json:"tags"`
}

func (r *UpdateRequest) toTags(resourceType string, resourceID uint) []*tagmodels.Tag {
	tags := make([]*tagmodels.Tag, 0)
	for _, tag := range r.Tags {
		tags = append(tags, &tagmodels.Tag{
			ResourceType: resourceType,
			ResourceID:   resourceID,
			Key:          tag.Key,
			Value:        tag.Value,
		})
	}
	return tags
}

func ofTags(tags []*tagmodels.Tag) *ListResponse {
	return &ListResponse{
		Tags: func() []*Tag {
			tagsResp := make([]*Tag, 0, len(tags))
			for _, tag := range tags {
				tagsResp = append(tagsResp, &Tag{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
			return tagsResp
		}(),
	}
}

type CreateMetatagsRequest struct {
	Metatags []*tagmodels.Metatag
}
