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
