package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Cluster struct {
	global.Model

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
}

type ClusterWithEnvAndRegion struct {
	*Cluster

	EnvironmentName   string
	RegionName        string
	RegionDisplayName string
}
