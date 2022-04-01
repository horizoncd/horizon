package models

import "gorm.io/gorm"

type K8SCluster struct {
	gorm.Model
	Name          string
	Server        string
	Certificate   string
	IngressDomain string
	CreatedBy     uint
	UpdatedBy     uint
	DeletedTs     int64
}

func (K8SCluster) TableName() string {
	return "tb_k8s_cluster"
}
