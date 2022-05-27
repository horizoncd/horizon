package models

import (
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Region struct {
	global.Model

	Name          string
	DisplayName   string
	Server        string
	Certificate   string
	IngressDomain string
	HarborID      uint `gorm:"column:harbor_id"`
	Disabled      bool
	CreatedBy     uint
	UpdatedBy     uint
}

// RegionEntity region entity, region with Harbor
type RegionEntity struct {
	*Region

	Harbor *harbormodels.Harbor
}

type RegionPart struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
}

type RegionParts []*RegionPart
