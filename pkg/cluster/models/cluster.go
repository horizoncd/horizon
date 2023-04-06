package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

type Cluster struct {
	global.Model

	ApplicationID   uint
	Name            string
	EnvironmentName string
	RegionName      string
	Description     string
	GitURL          string
	GitSubfolder    string
	GitRef          string
	GitRefType      string
	Template        string
	TemplateRelease string
	Status          string
	CreatedBy       uint
	UpdatedBy       uint
	ExpireSeconds   uint
}

type ClusterWithRegion struct {
	*Cluster

	RegionDisplayName string
}
