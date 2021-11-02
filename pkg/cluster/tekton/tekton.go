package tekton

import (
	"context"

	"g.hz.netease.com/horizon/pkg/config/tekton"
	"github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"

	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
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
	// PipelineRun 结构体用来接收接口参数，并传递给tekton trigger所暴露的接口
	PipelineRun struct {
		// 应用名称
		Application string `json:"application"`
		// 集群名称
		Cluster string `json:"cluster"`
		// 环境名称，overmind的环境名称，比如test、dev等
		Environment string `json:"environment"`
		// git相关参数
		Git PipelineRunGit `json:"git"`
		// build.xml的base64编码内容
		BuildXML string `json:"buildxml"`
		// dockerfile的base64编码内容
		Dockerfile PipelineRunDockerfile `json:"dockerfile"`
		// dockerfile的buildArgs
		// 由于tekton triggers不支持结构化数据，故该字段类型为字符串，格式为：
		// argName1::argValue1##argName2::argValue2...
		// （1）使用两个井号##分隔两个不同的参数；（2）使用两个冒号::分割一个参数的参数名和参数值
		BuildArgs string `json:"buildArgs"`
		// docker镜像的url
		ImageURL string `json:"imageurl"`
		// 操作人
		Operator string `json:"operator"`
	}
	PipelineRunGit struct {
		// git仓库的地址
		URL string `json:"url"`
		// git仓库的分支
		Branch string `json:"branch"`
		// git仓库的子目录
		Subfolder string `json:"subfolder"`
	}
	PipelineRunDockerfile struct {
		Content string `json:"content"`
		Path    string `json:"path"`
	}
)
