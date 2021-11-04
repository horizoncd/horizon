package log

import (
	"net/http"

	"github.com/tektoncd/cli/pkg/cli"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
)

// TektonParams 继承 https://github.com/tektoncd/cli/blob/v0.13.1/pkg/cli/params.go 的TektonParams，
// 用来获取pipelineRun的日志
type TektonParams struct {
	*cli.TektonParams
	namespace  string
	tekton     tektonclientset.Interface
	dynamic    dynamic.Interface
	kubeClient k8s.Interface
	clients    *cli.Clients
}

func NewTektonParams(dynamic dynamic.Interface, kubeClient k8s.Interface,
	tekton tektonclientset.Interface, namespace string) *TektonParams {
	return &TektonParams{
		namespace:  namespace,
		dynamic:    dynamic,
		kubeClient: kubeClient,
		tekton:     tekton,
	}
}

// 重写Clients方法，替换为我们自己的Clients
func (t *TektonParams) Clients() (*cli.Clients, error) {
	t.clients = &cli.Clients{
		Tekton:     t.tekton,
		Kube:       t.kubeClient,
		HTTPClient: http.Client{},
		Dynamic:    t.dynamic,
	}
	return t.clients, nil
}

// 重写Namespace方法，获取我们自己的Namespace
func (t *TektonParams) Namespace() string {
	return t.namespace
}
