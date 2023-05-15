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

package workload

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/core/operater"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

var abilities = make([]Workload, 0, 4)

var Resources = make([]operater.Resource, 0, 16)

func Register(ability Workload, gvrs ...schema.GroupVersionResource) {
	abilities = append(abilities, ability)
	gvrsUnderResource := make([]operater.Resource, 0, len(gvrs))
	for _, gvr := range gvrs {
		gvrsUnderResource = append(gvrsUnderResource, operater.Resource{
			GVR: gvr,
		})
	}
	Resources = append(Resources, gvrsUnderResource...)
}

type Handler func(workload Workload) bool

func LoopAbilities(handlers ...Handler) {
	for _, handler := range handlers {
		for _, ability := range abilities {
			if !handler(ability) {
				break
			}
		}
	}
}

type Workload interface {
	MatchGK(gk schema.GroupKind) bool
	Action(aName string, un *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type GreyscaleReleaser interface {
	Workload
	GetSteps(node *v1alpha1.ResourceNode, client *kube.Client) (*Step, error)
}

type PodsLister interface {
	Workload
	ListPods(node *v1alpha1.ResourceNode,
		factory dynamicinformer.DynamicSharedInformerFactory) ([]corev1.Pod, error)
}

type HealthStatusGetter interface {
	Workload
	IsHealthy(node *v1alpha1.ResourceNode, client *kube.Client) (bool, error)
}
