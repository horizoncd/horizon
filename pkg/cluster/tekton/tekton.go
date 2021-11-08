package tekton

import (
	"context"

	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	"g.hz.netease.com/horizon/pkg/config/tekton"

	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

type Interface interface {
	// GetLatestPipelineRun get latest pipelinerun
	GetLatestPipelineRun(ctx context.Context, application, cluster string) (*v1beta1.PipelineRun, error)
	// GetRunningPipelineRun get running pipelinerun
	GetRunningPipelineRun(ctx context.Context, application, cluster string) (*v1beta1.PipelineRun, error)
	// CreatePipelineRun create pipelinerun
	CreatePipelineRun(ctx context.Context, pr *PipelineRun) (string, error)
	// StopPipelineRun stop pipelinerun
	StopPipelineRun(ctx context.Context, application, cluster string) error
	// GetLatestPipelineRunLog get latest pipelinerun log
	GetLatestPipelineRunLog(ctx context.Context, application, cluster string) (<-chan log.Log, <-chan error, error)
	// GetPipelineRunLog 根据传入的pipelineRun获取该pipelineRun对应的log
	GetPipelineRunLog(ctx context.Context, pr *v1beta1.PipelineRun) (<-chan log.Log, <-chan error, error)
	// DeletePipelineRun 删除pipelineRun
	DeletePipelineRun(ctx context.Context, pr *v1beta1.PipelineRun) error
}

// Tekton 表示一个Tekton实例，其中包含它所对应的trigger server地址，生成资源的命名空间，
// 以及过滤日志所使用的filteredTasks和filteredSteps，以及所对应的clients
type Tekton struct {
	// tekton trigger serrver 的地址
	server string
	// tekton 资源的命名空间
	namespace string
	// 用来过滤日志
	filteredTasks string
	// 用来过滤日志
	filteredSteps string
	// clients，用来获取tekton资源、k8s资源等
	client *Client
}

// NewTekton 实例化一个Tekton实例
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
	// PipelineRun 结构体用来传递给tekton trigger所暴露的接口
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
	}
	PipelineRunGit struct {
		URL       string `json:"url"`
		Branch    string `json:"branch"`
		Subfolder string `json:"subfolder"`
		Commit    string `json:"commit"`
	}
)
