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
	"strings"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
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
