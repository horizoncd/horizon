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
	"context"
	"fmt"

	"github.com/horizoncd/horizon/pkg/util/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func ObjIntoPod(objs ...runtime.Object) []corev1.Pod {
	pods := make([]corev1.Pod, len(objs))
	for i, obj := range objs {
		pod := corev1.Pod{}
		err := ObjUnmarshal(obj, &pod)
		if err != nil {
			log.Errorf(context.TODO(), "failed to unmarshal pod object: %v", err)
			continue
		}
		pods[i] = pod
	}
	return pods
}

func ObjUnmarshal(obj runtime.Object, container interface{}) error {
	un, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("obj is not *unstructured.Unstructured")
	}

	return runtime.DefaultUnstructuredConverter.FromUnstructured(un.Object, container)
}
