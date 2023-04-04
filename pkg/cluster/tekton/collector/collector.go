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

package collector

import (
	"context"
	"strconv"

	"github.com/horizoncd/horizon/pkg/cluster/tekton/log"
	prmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/horizoncd/horizon/pkg/cluster/tekton/metrics"
	timeutil "github.com/horizoncd/horizon/pkg/util/time"
)

// Object the pipelinerun object to be collected
type Object struct {
	// Metadata meta data
	Metadata *ObjectMeta `json:"metadata"`
	// PipelineRun v1beta1.PipelineRun
	PipelineRun *v1beta1.PipelineRun `json:"pipelineRun"`
}

type (
	ObjectMeta struct {
		Application       string             `json:"application"`
		ApplicationID     string             `json:"applicationID"`
		Cluster           string             `json:"cluster"`
		ClusterID         string             `json:"clusterID"`
		Environment       string             `json:"environment"`
		Operator          string             `json:"operator"`
		CreationTimestamp string             `json:"creationTimestamp"`
		PipelineRun       *PipelineRunStatus `json:"pipelineRun"`
	}
	PipelineRunStatus struct {
		StatusMeta `json:",inline"`
		Pipeline   string `json:"pipeline"`
	}
	StatusMeta struct {
		Name            string       `json:"name"`
		Result          string       `json:"result"`
		DurationSeconds float64      `json:"durationSeconds"`
		StartTime       *metav1.Time `json:"startTime"`
		CompletionTime  *metav1.Time `json:"completionTime"`
	}
)

type Log struct {
	LogChannel <-chan log.Log
	ErrChannel <-chan error

	LogBytes []byte
}

func NewObjectMeta(horizonMetaData *global.HorizonMetaData, pr *v1beta1.PipelineRun) *ObjectMeta {
	wrappedPr := &metrics.WrappedPipelineRun{
		PipelineRun: pr,
	}
	prMetadata := wrappedPr.ResolveMetadata()
	prResult := wrappedPr.ResolvePrResult()
	return &ObjectMeta{
		Application:       horizonMetaData.Application,
		ApplicationID:     strconv.Itoa(int(horizonMetaData.ApplicationID)),
		Cluster:           horizonMetaData.Cluster,
		ClusterID:         strconv.Itoa(int(horizonMetaData.ClusterID)),
		Environment:       horizonMetaData.Environment,
		Operator:          horizonMetaData.Operator,
		CreationTimestamp: timeutil.K8sTimeToStrByNowTimezone(pr.CreationTimestamp),
		PipelineRun: &PipelineRunStatus{
			StatusMeta: StatusMeta{
				Name:            prMetadata.Name,
				Result:          prResult.Result,
				DurationSeconds: prResult.DurationSeconds,
				StartTime:       prResult.StartTime,
				CompletionTime:  prResult.CompletionTime,
			},
			Pipeline: prMetadata.Pipeline,
		},
	}
}

// nolint
// -package=mock_collector
//
//go:generate mockgen -source=$GOFILE -destination=../../../../mock/pkg/cluster/tekton/collector/collector_mock.go
type Interface interface {
	// Collect log & object for pipelinerun
	Collect(ctx context.Context, pr *v1beta1.PipelineRun, horizonMetaData *global.HorizonMetaData) (*CollectResult, error)

	// GetPipelineRunLog gets pipelinerun log from collector
	GetPipelineRunLog(ctx context.Context, pr *prmodels.Pipelinerun) (*Log, error)

	// GetPipelineRunObject get pipelinerun object from collector
	GetPipelineRunObject(ctx context.Context, object string) (*Object, error)

	// GetPipelineRun gets tekton pipelinerun
	GetPipelineRun(ctx context.Context, pr *prmodels.Pipelinerun) (*v1beta1.PipelineRun, error)
}

var _ Interface = (*S3Collector)(nil)

func resolveObjMetadata(pr *v1beta1.PipelineRun, horizonMetaData *global.HorizonMetaData) *ObjectMeta {
	return NewObjectMeta(horizonMetaData, pr)
}
