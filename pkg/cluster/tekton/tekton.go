package tekton

import (
	"context"

	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	"g.hz.netease.com/horizon/pkg/config/tekton"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type Interface interface {
	GetPipelineRunByID(ctx context.Context, cluster string,
		clusterID, pipelinerunID uint) (*v1beta1.PipelineRun, error)
	// CreatePipelineRun create pipelinerun
	CreatePipelineRun(ctx context.Context, pr *PipelineRun) (string, error)
	// StopPipelineRun stop pipelinerun
	StopPipelineRun(ctx context.Context, cluster string, clusterID, pipelinerunID uint) error
	GetPipelineRunLogByID(ctx context.Context, cluster string,
		clusterID, pipelinerunID uint) (<-chan log.Log, <-chan error, error)
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
	}
	PipelineRunGit struct {
		URL       string `json:"url"`
		Branch    string `json:"branch"`
		Tag       string `json:"tag"`
		Subfolder string `json:"subfolder"`
		Commit    string `json:"commit"`
	}
)
