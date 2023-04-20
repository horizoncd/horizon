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

package templateschematag

import (
	"github.com/horizoncd/horizon/pkg/authentication/user"
	tagmodels "github.com/horizoncd/horizon/pkg/templateschematag/models"
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

func (r *UpdateRequest) toClusterTemplateSchemaTags(clusterID uint,
	user user.User) []*tagmodels.ClusterTemplateSchemaTag {
	clusterSchemaTags := make([]*tagmodels.ClusterTemplateSchemaTag, 0)
	for _, tag := range r.Tags {
		clusterSchemaTags = append(clusterSchemaTags, &tagmodels.ClusterTemplateSchemaTag{
			ClusterID: clusterID,
			Key:       tag.Key,
			Value:     tag.Value,
			CreatedBy: user.GetID(),
			UpdatedBy: user.GetID(),
		})
	}
	return clusterSchemaTags
}

func ofClusterTemplateSchemaTags(clusterTemplateSchemaTags []*tagmodels.ClusterTemplateSchemaTag) *ListResponse {
	return &ListResponse{
		Tags: func() []*Tag {
			tags := make([]*Tag, 0, len(clusterTemplateSchemaTags))
			for _, tag := range clusterTemplateSchemaTags {
				tags = append(tags, &Tag{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
			return tags
		}(),
	}
}
