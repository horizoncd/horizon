package service

import (
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
)

// ClusterDetail contains the fullPath
type ClusterDetail struct {
	clustermodels.Cluster
	FullPath string
}
