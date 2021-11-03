package models

import "gorm.io/gorm"

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
	EnvironmentRegionID uint
	CreatedBy           uint
	UpdatedBy           uint
}

type ClusterWithEnvAndRegion struct {
	*Cluster

	EnvironmentName   string
	RegionName        string
	RegionDisplayName string
}
