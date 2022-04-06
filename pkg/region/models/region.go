package models

import (
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	"g.hz.netease.com/horizon/pkg/k8scluster/models"
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Region struct {
	global.Model

	Name         string
	DisplayName  string
	K8SClusterID uint `gorm:"column:k8s_cluster_id"`
	HarborID     uint `gorm:"column:harbor_id"`
	CreatedBy    uint
	UpdatedBy    uint
}

// RegionEntity region entity, region with its k8sCluster and Harbor
type RegionEntity struct {
	*Region

	K8SCluster *models.K8SCluster
	Harbor     *harbormodels.Harbor
}
