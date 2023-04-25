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

package tekton

import (
	"os"
	"testing"

	tektonconfig "github.com/horizoncd/horizon/pkg/config/tekton"

	"github.com/stretchr/testify/assert"
)

func TestNewTekton(t *testing.T) {
	tektonConfig := &tektonconfig.Tekton{
		Kubeconfig: "/",
	}
	tekton, err := NewTekton(tektonConfig)
	assert.Nil(t, tekton)
	assert.NotNil(t, err)

	tektonConfig = &tektonconfig.Tekton{
		Kubeconfig: "",
	}
	tekton, err = NewTekton(tektonConfig)
	host, port := os.Getenv("KUBERNETES_SERVICE_HOST"), os.Getenv("KUBERNETES_SERVICE_PORT")
	if len(host) == 0 || len(port) == 0 {
		assert.Nil(t, tekton)
		assert.NotNil(t, err)
	} else {
		assert.NotNil(t, tekton)
		assert.Nil(t, err)
	}
}
