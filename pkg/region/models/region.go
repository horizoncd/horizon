package models

import (
	"g.hz.netease.com/horizon/pkg/k8scluster/models"

	"gorm.io/gorm"
)

type Region struct {
	gorm.Model

	Region       string
	K8SClusterID uint
	CreatedBy    string
	UpdatedBy    string
}

type RegionWithK8SCluster struct {
	gorm.Model

	Region     string
	K8SCluster *models.K8SCluster
	CreatedBy  string
	UpdatedBy  string
}
