package workload

import (
	"container/heap"

	"github.com/horizoncd/horizon/pkg/util/kube"
	"github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	corev1 "k8s.io/api/core/v1"
)

var Abilities = make(PriorityQueue, 0, 4)

func Register(ability Workload, priority uint) {
	workload := WorkloadWithPriority{
		workload: ability,
		index:    1,
		priority: priority,
	}
	heap.Push(&Abilities, &workload)
}

type Handler func(workload Workload) bool

func LoopAbilities(handlers ...Handler) {
	for _, handler := range handlers {
		for _, ability := range Abilities {
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

type PriorityQueue []*WorkloadWithPriority

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the lowest, not lowest, priority so we use less than here.
	return pq[i].priority < pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*WorkloadWithPriority)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}
