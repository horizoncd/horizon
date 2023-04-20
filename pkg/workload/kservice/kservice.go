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

package kservice

import (
	"context"
	"strconv"

	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/workload"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/polymorphichelpers"
	servicev1 "knative.dev/serving/pkg/apis/serving/v1"
)

func init() {
	workload.Register(ability)
}

// please refer to github.com/horizoncd/horizon/pkg/cluster/cd/workload/workload.go
var ability = &service{}

type service struct{}

func (*service) MatchGK(gk schema.GroupKind) bool {
	return gk.Group == "serving.knative.dev" && gk.Kind == "Service"
}

func (*service) getServiceByNode(node *v1alpha1.ResourceNode, client *kube.Client) (*servicev1.Service, error) {
	gvr := schema.GroupVersionResource{
		Group:    "serving.knative.dev",
		Version:  node.Version,
		Resource: "services",
	}

	un, err := client.Dynamic.Resource(gvr).Namespace(node.Namespace).
		Get(context.TODO(), node.Name, metav1.GetOptions{})
	if err != nil {
		return nil, perror.Wrapf(
			herrors.NewErrGetFailed(herrors.ResourceInK8S,
				"failed to get deployment in k8s"),
			"failed to get deployment in k8s: deployment = %s, err = %v", node.Name, err)
	}

	var ksvc *servicev1.Service
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(un.UnstructuredContent(), &ksvc)
	if err != nil {
		return nil, err
	}
	return ksvc, nil
}

func (s *service) ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error) {
	instance, err := s.getServiceByNode(node, client)
	if err != nil {
		return nil, err
	}

	selector := metav1.FormatLabelSelector(&metav1.LabelSelector{MatchLabels: instance.Spec.Template.Labels})
	pods, err := client.Basic.CoreV1().Pods(instance.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: selector, ResourceVersion: "0"})
	if err != nil {
		return nil, err
	}

	return pods.Items, nil
}

func (s *service) IsHealthy(node *v1alpha1.ResourceNode,
	client *kube.Client) (bool, error) {
	instance, err := s.getServiceByNode(node, client)
	if err != nil {
		return true, err
	}

	if instance == nil {
		return true, nil
	}

	labels := polymorphichelpers.MakeLabels(instance.Spec.Template.ObjectMeta.Labels)
	pods, err := client.Basic.CoreV1().Pods(instance.Namespace).
		List(context.TODO(), metav1.ListOptions{LabelSelector: labels, ResourceVersion: "0"})
	if err != nil {
		return true, err
	}
	log.Debugf(context.TODO(), "[knative service: %v] pods count = %v", instance.Name, len(pods.Items))

	annos := instance.Spec.Template.ObjectMeta.Annotations
	min, _ := strconv.Atoi(annos["autoscaling.knative.dev/minScale"])
	max, _ := strconv.Atoi(annos["autoscaling.knative.dev/maxScale"])
	log.Debugf(context.TODO(), "[knative service: %v] minScale = %v, maxScale = %v", instance.Name, min, max)

	count := 0

OUTTER:
	for _, pod := range pods.Items {
		m := make(map[string]string)
		for _, container := range pod.Spec.Containers {
			m[container.Name] = container.Image
		}

		for _, container := range instance.Spec.Template.Spec.Containers {
			if image := m[container.Name]; image != container.Image {
				log.Debugf(context.TODO(),
					"[knative service: %v] expect container(%v)'s image = %v, pod(%v) container(%v)'s image = %v",
					instance.Name, container.Name, image, pod.Name, container.Name, container.Image)
				continue OUTTER
			}
		}

		for k, v := range instance.Spec.Template.ObjectMeta.Annotations {
			if pod.Annotations[k] != v {
				log.Debugf(context.TODO(),
					"[knative service: %v] expect annotation(%v) = %v, pod(%v) annotation(%v) = %v",
					instance.Name, k, v,
					pod.Name, k, pod.Annotations[k])
				continue OUTTER
			}
		}
		count++
		if count > max {
			break
		}
	}
	log.Debugf(context.TODO(),
		"[knative service: %v] count = %v, min = %v, max = %v",
		instance.Name, count, min, max)
	return count >= min && count <= max, nil
}

func (*service) Action(actionName string, un *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return un, nil
}
