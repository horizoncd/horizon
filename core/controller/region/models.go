package region

import (
	"time"

	"g.hz.netease.com/horizon/pkg/region/models"
)

type Region struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	DisplayName   string    `json:"displayName"`
	Server        string    `json:"server"`
	Certificate   string    `json:"certificate"`
	IngressDomain string    `json:"ingressDomain"`
	HarborID      uint      `json:"harborID"`
	HarborName    string    `json:"harborName"`
	HarborServer  string    `json:"harborServer"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
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
}

func ofRegionEntities(entities []*models.RegionEntity) []*Region {
	regions := make([]*Region, 0)

	for _, entity := range entities {
		regions = append(regions, &Region{
			ID:            entity.ID,
			Name:          entity.Name,
			DisplayName:   entity.DisplayName,
			Server:        entity.Server,
			IngressDomain: entity.IngressDomain,
			HarborID:      entity.Harbor.ID,
			HarborName:    entity.Harbor.Name,
			HarborServer:  entity.Harbor.Server,
			CreatedAt:     entity.CreatedAt,
			UpdatedAt:     entity.UpdatedAt,
		})
	}

	return regions
}
