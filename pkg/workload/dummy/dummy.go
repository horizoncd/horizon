package dummy

import (
	"github.com/horizoncd/horizon/pkg/workload"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func init() {
	workload.Register(ability,
		schema.GroupVersionResource{
			Group:    "argoproj.io",
			Version:  "v1alpha1",
			Resource: "applications",
		},
		schema.GroupVersionResource{
			Group:    "apps",
			Version:  "v1",
			Resource: "deployments",
		})
}

var ability = &dummy{}

// dummy is a dummy implementation of workload.Interface
// It is used to register the ability to handle some required resources.
type dummy struct{}

func (*dummy) MatchGK(_ schema.GroupKind) bool {
	return false
}

func (*dummy) Action(_ string, un *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return un, nil
}
