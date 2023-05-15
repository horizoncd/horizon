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

package getter

import (
	"fmt"
	"reflect"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/workload"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

type Helper struct {
	inner workload.Workload
}

func New(inner workload.Workload) *Helper {
	return &Helper{inner: inner}
}

func (w *Helper) GetSteps(node *v1alpha1.ResourceNode, client *kube.Client) (*workload.Step, error) {
	releaser, ok := w.inner.(workload.GreyscaleReleaser)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support greyscale release", reflect.TypeOf(w.inner))
	}
	steps, err := releaser.GetSteps(node, client)
	if err != nil {
		return nil, herrors.NewErrGetFailed(herrors.ResourceInK8S,
			fmt.Sprintf("failed to get steps: resource name = %v, err = %v", node.Name, err))
	}
	return steps, nil
}

// TODO: remove this after using informer
func (w *Helper) ListPods(node *v1alpha1.ResourceNode,
	factory dynamicinformer.DynamicSharedInformerFactory) ([]corev1.Pod, error) {
	lister, ok := w.inner.(workload.PodsLister)
	if !ok {
		return nil, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support list release", reflect.TypeOf(w.inner))
	}
	pods, err := lister.ListPods(node, factory)
	if err != nil {
		return nil,
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to list pods: resource name = %v, err = %v", node.Name, err))
	}
	return pods, nil
}

func (w *Helper) IsHealthy(node *v1alpha1.ResourceNode, client *kube.Client) (bool, error) {
	statusGetter, ok := w.inner.(workload.HealthStatusGetter)
	if !ok {
		return true, perror.Wrapf(herrors.ErrMethodNotImplemented,
			"workload %v not support get health status", reflect.TypeOf(w.inner))
	}

	isHealthy, err := statusGetter.IsHealthy(node, client)
	if err != nil {
		return isHealthy,
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to get healthy: resource name = %v, err = %v", node.Name, err))
	}
	return isHealthy, nil
}
