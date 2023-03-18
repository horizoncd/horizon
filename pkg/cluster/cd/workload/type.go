package workload

import (
	"k8s.io/api/core/v1"
)

type Step struct {
	Index        int
	Total        int
	Replicas     []int
	ManualPaused bool
}

type Revision struct {
	Name string
	Pods []v1.Pod
}
