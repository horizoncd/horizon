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
