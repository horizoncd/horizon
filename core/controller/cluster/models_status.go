package cluster

import (
	"github.com/horizoncd/horizon/pkg/grafana"
	v1 "k8s.io/api/core/v1"
)

type GetClusterStatusResponse struct {
	RunningTask       *RunningTask       `json:"runningTask" yaml:"runningTask"`
	LatestPipelinerun *LatestPipelinerun `json:"latestPipelinerun,omitempty"`
	ClusterStatus     interface{}        `json:"clusterStatus,omitempty" yaml:"clusterStatus,omitempty"`
	TTLSeconds        *uint              `json:"ttlSeconds,omitempty" yaml:"ttlSeconds,omitempty"`
}

// RunningTask the most recent task in progress
type RunningTask struct {
	Task string `json:"task" yaml:"task"`
	// the latest buildDeploy pipelinerun ID
	PipelinerunID uint   `json:"pipelinerunID,omitempty"`
	TaskStatus    string `json:"taskStatus,omitempty" yaml:"taskStatus,omitempty"`
}

// LatestPipelinerun latest pipelinerun
type LatestPipelinerun struct {
	ID     uint   `json:"id"`
	Action string `json:"action"`
}

type GetGrafanaDashboardsResponse struct {
	Host       string               `json:"host"`
	Params     map[string]string    `json:"params"`
	Dashboards []*grafana.Dashboard `json:"dashboards"`
}

type GetClusterPodResponse struct {
	v1.Pod
}
