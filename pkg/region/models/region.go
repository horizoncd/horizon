package models

import (
	"g.hz.netease.com/horizon/pkg/k8scluster/models"

	"gorm.io/gorm"
)

type Region struct {
	gorm.Model

	Name         string
	DisplayName  string
	K8SClusterID uint
	CreatedBy    uint
	UpdatedBy    uint
}

type RegionWithK8SCluster struct {
	gorm.Model

	Name        string
	DisplayName string
	K8SCluster  *models.K8SCluster
	CreatedBy   uint
	UpdatedBy   uint
}
