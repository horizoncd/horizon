package models

import (
	registrymodels "g.hz.netease.com/horizon/pkg/registry/models"
	"g.hz.netease.com/horizon/pkg/server/global"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
)

type Region struct {
	global.Model

	Name          string
	DisplayName   string
	Server        string
	Certificate   string
	IngressDomain string
	PrometheusURL string
	RegistryID    uint `gorm:"column:registry_id"`
	Disabled      bool
	CreatedBy     uint
	UpdatedBy     uint
}

// RegionEntity region entity, region with Harbor
type RegionEntity struct {
	*Region

	Registry *registrymodels.Registry
	Tags     []*tagmodels.Tag
}

type RegionPart struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Disabled    bool   `json:"disabled"`
	IsDefault   bool   `json:"isDefault"`
}

type RegionParts []*RegionPart
