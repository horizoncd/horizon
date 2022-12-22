package workload

import (
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/horizoncd/horizon/pkg/util/kube"
	corev1 "k8s.io/api/core/v1"
)

var abilities = make([]*WorkloadWithPriority, 0, 4)

func Register(ability Workload) {
	workload := WorkloadWithPriority{
		workload: ability,
		index:    1,
	}
	abilities = append(abilities, &workload)
}

type Handler func(workload Workload) bool

func LoopAbilities(handlers ...Handler) {
	for _, handler := range handlers {
		for _, ability := range abilities {
			if !handler(ability.workload) {
				break
			}
		}
	}
}

type Workload interface {
	MatchGK(gk string) bool
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

// nolint
type WorkloadWithPriority struct {
	workload Workload

	index    int
	priority uint
}
