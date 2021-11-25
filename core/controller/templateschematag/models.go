package templateschematag

import (
	"g.hz.netease.com/horizon/pkg/authentication/user"
	clustertagmodels "g.hz.netease.com/horizon/pkg/templateschematag/models"
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
	user user.User) []*clustertagmodels.ClusterTemplateSchemaTag {
	clusterSchemaTags := make([]*clustertagmodels.ClusterTemplateSchemaTag, 0)
	for _, tag := range r.Tags {
		clusterSchemaTags = append(clusterSchemaTags, &clustertagmodels.ClusterTemplateSchemaTag{
			ClusterID: clusterID,
			Key:       tag.Key,
			Value:     tag.Value,
			CreatedBy: user.GetID(),
			UpdatedBy: user.GetID(),
		})
	}
	return clusterSchemaTags
}

func ofClusterTemplateSchemaTags(clusterTemplateSchemaTags []*clustertagmodels.ClusterTemplateSchemaTag) *ListResponse {
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
