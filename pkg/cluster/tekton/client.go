package tekton

import (
	"strings"

	"g.hz.netease.com/horizon/pkg/util/kube"
	"github.com/pkg/errors"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// Client 组装需要用到的Client
type Client struct {
	// 用来获取Tekton相关CRD的Client
	Tekton tektonclientset.Interface
	// 获取k8s核心资源的Client
	Kube k8s.Interface
	// 通过REST API方法访问k8s api的Client。Dynamic client可以获取k8s所有资源，包括核心资源和第三方资源
	Dynamic dynamic.Interface
}

// InitClient 初始化Clients
// Kubeconfig表示kubeconfig文件的地址。如果该地址为空，则默认使用InClusterConfig，即本Pod所在集群的config
func InitClient(kubeconfig string) (*Client, error) {
	var config *rest.Config
	var err error

	// startWith "/", kubeconfig is a filePath, else kubeconfig is fileContent
	if strings.HasPrefix(kubeconfig, "/") || len(kubeconfig) == 0 {
		config, _, err = kube.BuildClient(kubeconfig)
	} else {
		config, _, err = kube.BuildClientFromContent(kubeconfig)
	}
	if err != nil {
		return nil, err
	}

	tekton, err := tektonClient(config)
	if err != nil {
		return nil, err
	}
	kube, err := kubeClient(config)
	if err != nil {
		return nil, err
	}
	dynamic, err := dynamicClient(config)
	if err != nil {
		return nil, err
	}
	return &Client{
		Tekton:  tekton,
		Kube:    kube,
		Dynamic: dynamic,
	}, nil
}

func tektonClient(config *rest.Config) (tektonclientset.Interface, error) {
	cs, err := tektonclientset.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return cs, nil
}

func kubeClient(config *rest.Config) (k8s.Interface, error) {
	k8scs, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create k8s client from config")
	}
	return k8scs, nil
}

func dynamicClient(config *rest.Config) (dynamic.Interface, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create dynamic client from config")
	}
	return dynamicClient, err
}
