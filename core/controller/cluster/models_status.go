package cluster

type GetClusterStatusResponse struct {
	RunningTask   *RunningTask `json:"runningTask" yaml:"runningTask"`
	ClusterStatus interface{}  `json:"clusterStatus,omitempty" yaml:"clusterStatus,omitempty"`
}

// RunningTask 最近一次在执行中的Task
type RunningTask struct {
	Task       string `json:"task" yaml:"task"`
	TaskStatus string `json:"taskStatus,omitempty" yaml:"taskStatus,omitempty"`
}
