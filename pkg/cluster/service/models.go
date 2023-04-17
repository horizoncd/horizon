package service

import (
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
)

// ClusterDetail contains the fullPath.
type ClusterDetail struct {
	clustermodels.Cluster
	FullPath string
}
