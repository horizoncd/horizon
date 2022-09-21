package cluster

import (
	v1 "k8s.io/api/core/v1"
)

type GetClusterStatusResponse struct {
	RunningTask       *RunningTask       `json:"runningTask" yaml:"runningTask"`
	LatestPipelinerun *LatestPipelinerun `json:"latestPipelinerun,omitempty"`
	ClusterStatus     interface{}        `json:"clusterStatus,omitempty" yaml:"clusterStatus,omitempty"`
}

// RunningTask 最近一次在执行中的Task
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

type GetDashboardResponse struct {
	Basic      string `json:"basic" yaml:"basic"`
	Container  string `json:"container" yaml:"container"`
	Serverless string `json:"serverless,omitempty" yaml:"serverless,omitempty"`
	Memcached  string `json:"memcached,omitempty" yaml:"memcached,omitempty"`
}

type GrafanaDashboard struct {
	URL   string `json:"url"`
	Title string `json:"title"`
}

type GetClusterPodsResponse struct {
	Pods []KubePodInfo `json:"pods" yaml:"pods"`
}

type GetClusterPodResponse struct {
	v1.Pod
}

type KubePodInfo struct {
	Pod       string `json:"pod"`
	Container string `json:"container"`
	IP        string `json:"pod_ip"`
}

type QueryPodsSeriesResult struct {
	Status string        `json:"status"`
	Data   []KubePodInfo `json:"data"`
}
