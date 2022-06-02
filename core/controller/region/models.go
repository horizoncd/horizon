package region

import (
	"time"

	"g.hz.netease.com/horizon/core/controller/harbor"
	"g.hz.netease.com/horizon/core/controller/tag"
	"g.hz.netease.com/horizon/pkg/region/models"
)

type Region struct {
	ID            uint          `json:"id"`
	Name          string        `json:"name"`
	DisplayName   string        `json:"displayName"`
	Server        string        `json:"server"`
	Certificate   string        `json:"certificate"`
	IngressDomain string        `json:"ingressDomain"`
	Disabled      bool          `json:"disabled"`
	HarborID      uint          `json:"harborID"`
	Harbor        harbor.Harbor `json:"harbor"`
	Tags          []tag.Tag     `json:"tags"`
	CreatedAt     time.Time     `json:"createdAt"`
	UpdatedAt     time.Time     `json:"updatedAt"`
}

type CreateRegionRequest struct {
	Name          string
	DisplayName   string
	Server        string
	Certificate   string
	IngressDomain string
	HarborID      uint `json:"harborID"`
}

type UpdateRegionRequest struct {
	DisplayName   string
	Server        string
	Certificate   string
	IngressDomain string
	HarborID      uint `json:"harborID"`
	Disabled      bool
}

func ofRegionEntity(entity *models.RegionEntity) *Region {
	var tags []tag.Tag
	for _, t := range entity.Tags {
		tags = append(tags, tag.Tag{
			Key:   t.Key,
			Value: t.Value,
		})
	}
	r := &Region{
		ID:            entity.ID,
		Name:          entity.Name,
		DisplayName:   entity.DisplayName,
		Server:        entity.Server,
		IngressDomain: entity.IngressDomain,
		Certificate:   entity.Certificate,
		Disabled:      entity.Disabled,
		HarborID:      entity.HarborID,
		Tags:          tags,
		CreatedAt:     entity.CreatedAt,
		UpdatedAt:     entity.UpdatedAt,
	}
	if entity.Harbor != nil {
		r.Harbor = harbor.Harbor{
			Name: entity.Harbor.Name,
		}
	}
	return r
}

func ofRegionEntities(entities []*models.RegionEntity) []*Region {
	regions := make([]*Region, 0)

	for _, entity := range entities {
		regions = append(regions, ofRegionEntity(entity))
	}

	return regions
}
