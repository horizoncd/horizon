package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type K8SCluster struct {
	global.Model
	Name          string
	Server        string
	Certificate   string
	IngressDomain string
	CreatedBy     uint
	UpdatedBy     uint
}

func (K8SCluster) TableName() string {
	return "tb_k8s_cluster"
}
