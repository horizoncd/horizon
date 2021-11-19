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
}

func (K8SCluster) TableName() string {
	return "k8s_cluster"
}
