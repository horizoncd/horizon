package workload

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var abilities = make([]Workload, 0, 4)

func Register(ability Workload) {
	abilities = append(abilities, ability)
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
	ListPods(node *v1alpha1.ResourceNode, client *kube.Client) ([]corev1.Pod, error)
}

type HealthStatusGetter interface {
	Workload
	IsHealthy(node *v1alpha1.ResourceNode, client *kube.Client) (bool, error)
}
