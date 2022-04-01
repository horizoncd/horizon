package models

import (
	"gorm.io/gorm"
)

type Cluster struct {
	gorm.Model

	ApplicationID       uint
	Name                string
	Description         string
	GitURL              string
	GitSubfolder        string
	GitBranch           string
	Template            string
	TemplateRelease     string
	Status              string
	EnvironmentRegionID uint
	CreatedBy           uint
	UpdatedBy           uint
	DeletedTs           int64
}

type ClusterWithEnvAndRegion struct {
	*Cluster

	EnvironmentName   string
	RegionName        string
	RegionDisplayName string
}
