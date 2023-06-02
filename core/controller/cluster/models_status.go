// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/pkg/grafana"
	corev1 "k8s.io/api/core/v1"
)

type Step struct {
	Index    int   `json:"index"`
	Total    int   `json:"total"`
	Replicas []int `json:"replicas"`
}

type Revision struct {
	Pods map[string]interface{} `json:"pods"`
}

type StatusResponseV2 struct {
	Status string `json:"status"`
}

type PipelinerunStatusResponse struct {
	LatestPipelinerun *LatestPipelinerun `json:"latestPipelinerun,omitempty"`
	RunningTask       *RunningTask       `json:"runningTask,omitempty"`
}

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
	corev1.Pod
}
type ResourceNode struct {
	v1alpha1.ResourceNode
	PodDetail interface{} `json:"podDetail,omitempty"`
}

type GetResourceTreeResponse struct {
	Nodes map[string]*ResourceNode `json:"nodes"`
}

type GetStepResponse struct {
	Index        int   `json:"index"`
	Total        int   `json:"total"`
	Replicas     []int `json:"replicas"`
	ManualPaused bool  `json:"manualPaused"`
	AutoPromote  bool  `json:"autoPromote"`
}
