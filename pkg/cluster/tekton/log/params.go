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

package log

import (
	"net/http"

	"github.com/tektoncd/cli/pkg/cli"
	tektonclientset "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"
	"k8s.io/client-go/dynamic"
	k8s "k8s.io/client-go/kubernetes"
)

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

func (t *TektonParams) Clients() (*cli.Clients, error) {
	t.clients = &cli.Clients{
		Tekton:     t.tekton,
		Kube:       t.kubeClient,
		HTTPClient: http.Client{},
		Dynamic:    t.dynamic,
	}
	return t.clients, nil
}

func (t *TektonParams) Namespace() string {
	return t.namespace
}
