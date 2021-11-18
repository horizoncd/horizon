package clustertag

import (
	"g.hz.netease.com/horizon/pkg/authentication/user"
	clustertagmodels "g.hz.netease.com/horizon/pkg/clustertag/models"
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

func (r *UpdateRequest) toClusterTags(clusterID uint, user user.User) []*clustertagmodels.ClusterTag {
	clusterTags := make([]*clustertagmodels.ClusterTag, 0)
	for _, tag := range r.Tags {
		clusterTags = append(clusterTags, &clustertagmodels.ClusterTag{
			ClusterID: clusterID,
			Key:       tag.Key,
			Value:     tag.Value,
			CreatedBy: user.GetID(),
			UpdatedBy: user.GetID(),
		})
	}
	return clusterTags
}

func ofClusterTags(clusterTags []*clustertagmodels.ClusterTag) *ListResponse {
	return &ListResponse{
		Tags: func() []*Tag {
			tags := make([]*Tag, 0, len(clusterTags))
			for _, tag := range clusterTags {
				tags = append(tags, &Tag{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
			return tags
		}(),
	}
}
