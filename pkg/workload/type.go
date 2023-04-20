package workload

import (
	v1 "k8s.io/api/core/v1"
)

type Step struct {
	Index        int
	Total        int
	Replicas     []int
	ManualPaused bool
	AutoPromote  bool
}

type Revision struct {
	Name string
	Pods []v1.Pod
}
