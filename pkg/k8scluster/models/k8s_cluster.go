package models

import "gorm.io/gorm"

type K8SCluster struct {
	gorm.Model
	Name         string
	Certificate  string
	DomainSuffix string
	CreatedBy    string
	UpdatedBy    string
}

func (K8SCluster) TableName() string {
	return "k8s_cluster"
}
