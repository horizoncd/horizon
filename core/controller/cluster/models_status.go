package cluster

type GetClusterStatusResponse struct {
	RunningTask   *RunningTask `json:"runningTask" yaml:"runningTask"`
	ClusterStatus interface{}  `json:"clusterStatus,omitempty" yaml:"clusterStatus,omitempty"`
}

// RunningTask 最近一次在执行中的Task
type RunningTask struct {
	Task          string `json:"task" yaml:"task"`
	PipelinerunID uint   `json:"pipelinerunID,omitempty"`
	TaskStatus    string `json:"taskStatus,omitempty" yaml:"taskStatus,omitempty"`
}

type GetDashboardResponse struct {
	Basic      string `json:"basic" yaml:"basic"`
	Serverless string `json:"serverless,omitempty" yaml:"serverless,omitempty"`
}
