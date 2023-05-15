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

package deployment

import (
	"context"
	"fmt"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/workload"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
)

var (
	GVRDeployment = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	GVRReplicaSet = schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "replicasets",
	}
	GVRPod = schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "pods",
	}
)

func init() {
	workload.Register(ability, GVRDeployment, GVRReplicaSet, GVRPod)
}

// please refer to github.com/horizoncd/horizon/pkg/cluster/cd/workload/workload.go
var ability = &deployment{}

type deployment struct{}

func (*deployment) MatchGK(gk schema.GroupKind) bool {
	return gk.Group == "apps" && gk.Kind == "Deployment"
}

func (*deployment) getDeploy(node *v1alpha1.ResourceNode,
	factory dynamicinformer.DynamicSharedInformerFactory) (*v1.Deployment, error) {
	obj, err := factory.ForResource(GVRDeployment).Lister().ByNamespace(node.Namespace).Get(node.Name)
	if err != nil {
		return nil,
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to get deployment in k8s: deployment = %s, ns = %v, err = %v",
					node.Name, node.Namespace, err),
			)
	}
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return nil,
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to convert obj into unstructured: name = %s, ns = %v",
					node.Name, node.Namespace),
			)
	}
	deploy := &v1.Deployment{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, deploy)
	if err != nil {
		return nil,
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				fmt.Sprintf("failed to convert unstructured into deployment: name = %s, ns = %v, err = %v",
					node.Name, node.Namespace, err),
			)
	}
	return deploy, nil
}

func (*deployment) getDeployByNode(node *v1alpha1.ResourceNode, client *kube.Client) (*v1.Deployment, error) {
	deploy, err := client.Basic.AppsV1().Deployments(node.Namespace).Get(context.TODO(), node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				"failed to get deployment in k8s"),
			"failed to get deployment in k8s: deployment = %s, ns = %v, err = %v", node.Name, node.Namespace, err)
	}
	return deploy, nil
}

func (d *deployment) IsHealthy(node *v1alpha1.ResourceNode,
	client *kube.Client) (bool, error) {
	instance, err := d.getDeployByNode(node, client)
	if err != nil {
		return true, err
	}

	if instance.Status.ObservedGeneration != instance.Generation {
		return false, nil
	}

	return instance.Status.AvailableReplicas == *instance.Spec.Replicas, nil
}

func (d *deployment) ListPods(node *v1alpha1.ResourceNode,
	factory dynamicinformer.DynamicSharedInformerFactory) ([]corev1.Pod, error) {
	instance, err := d.getDeploy(node, factory)
	if err != nil {
		return nil, err
	}

	selector := labels.SelectorFromSet(instance.Spec.Selector.MatchLabels)
	objs, err := factory.ForResource(GVRPod).Lister().ByNamespace(node.Namespace).List(selector)
	if err != nil {
		return nil, err
	}

	pods := workload.ObjIntoPod(objs...)

	return pods, nil
}

func (*deployment) Action(actionName string, un *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return un, nil
}
