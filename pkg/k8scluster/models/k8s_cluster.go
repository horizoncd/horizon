package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type K8SCluster struct {
	gorm.Model
	Name          string
	Server        string
	Certificate   string
	IngressDomain string
	CreatedBy     uint
	UpdatedBy     uint
	DeletedTs     soft_delete.DeletedAt
}

func (K8SCluster) TableName() string {
	return "tb_k8s_cluster"
}
