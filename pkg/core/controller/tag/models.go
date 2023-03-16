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
