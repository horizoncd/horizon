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

package tekton

import (
	"context"

	"github.com/horizoncd/horizon/pkg/cluster/tekton/log"
	"github.com/horizoncd/horizon/pkg/config/tekton"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/tekton/tekton_mock.go -package=mock_tekton
type Interface interface {
	GetPipelineRunByID(ctx context.Context, ciEventID string) (*v1beta1.PipelineRun, error)
	// CreatePipelineRun create pipelinerun
	CreatePipelineRun(ctx context.Context, pr *PipelineRun) (string, error)
	// StopPipelineRun stop pipelinerun
	StopPipelineRun(ctx context.Context, ciEventID string) error
	GetPipelineRunLogByID(ctx context.Context, ciEventID string) (<-chan log.Log, <-chan error, error)
	GetPipelineRunLog(ctx context.Context, pr *v1beta1.PipelineRun) (<-chan log.Log, <-chan error, error)
	DeletePipelineRun(ctx context.Context, pr *v1beta1.PipelineRun) error
}

type Tekton struct {
	server    string
	namespace string
	client    *Client
}

func NewTekton(tektonConfig *tekton.Tekton) (*Tekton, error) {
	client, err := InitClient(tektonConfig.Kubeconfig)
	if err != nil {
		return nil, err
	}
	return &Tekton{
		server:    tektonConfig.Server,
		namespace: tektonConfig.Namespace,
		client:    client,
	}, nil
}

type (
	PipelineRun struct {
		Application      string                 `json:"application"`
		ApplicationID    uint                   `json:"applicationID"`
		Cluster          string                 `json:"cluster"`
		ClusterID        uint                   `json:"clusterID"`
		Environment      string                 `json:"environment"`
		Git              PipelineRunGit         `json:"git"`
		ImageURL         string                 `json:"imageURL"`
		Operator         string                 `json:"operator"`
		PipelinerunID    uint                   `json:"pipelinerunID"`
		PipelineJSONBlob map[string]interface{} `json:"pipelineJSONBlob"`
		Region           string                 `json:"region"`
		RegionID         uint                   `json:"regionID"`
		Template         string                 `json:"template"`
		Token            string                 `json:"token"`
	}
	PipelineRunGit struct {
		URL       string `json:"url"`
		Branch    string `json:"branch"`
		Tag       string `json:"tag"`
		Subfolder string `json:"subfolder"`
		Commit    string `json:"commit"`
	}
)
