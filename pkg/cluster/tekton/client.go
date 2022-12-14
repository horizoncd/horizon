package tekton

import (
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/util/kube"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type Client struct {
	Tekton  tektonclientset.Interface
	Kube    k8s.Interface
	Dynamic dynamic.Interface
}

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
		return nil, herrors.NewErrCreateFailed(herrors.TektonClient, err.Error())
	}
	return cs, nil
}

func kubeClient(config *rest.Config) (k8s.Interface, error) {
	k8scs, err := k8s.NewForConfig(config)
	if err != nil {
		return nil, herrors.NewErrCreateFailed(herrors.K8SClient, err.Error())
	}
	return k8scs, nil
}

func dynamicClient(config *rest.Config) (dynamic.Interface, error) {
	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, herrors.NewErrCreateFailed(herrors.K8SClient, err.Error())
	}
	return dynamicClient, err
}
