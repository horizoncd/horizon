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

package argocd

import (
	"testing"

	"github.com/horizoncd/horizon/pkg/config/argocd"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	argoCDMapper := make(map[string]*argocd.ArgoCD)
	argoCDTest := &argocd.ArgoCD{
		URL:       "http://test.argo.com",
		Token:     "token1",
		Namespace: "argocd",
	}
	argoCDReg := &argocd.ArgoCD{
		URL:       "http://reg.argo.com",
		Token:     "token1",
		Namespace: "argocd",
	}
	argoCDMapper["test"] = argoCDTest
	argoCDMapper["reg"] = argoCDReg
	factory := NewFactory(argoCDMapper)
	assert.NotNil(t, factory)

	argoCD, err := factory.GetArgoCD("test")
	assert.Nil(t, err)
	assert.NotNil(t, argoCD)
	assert.Equal(t, argoCD, NewArgoCD(argoCDTest.URL, argoCDTest.Token, argoCDTest.Namespace))

	argoCD, err = factory.GetArgoCD("reg")
	assert.Nil(t, err)
	assert.NotNil(t, argoCD)
	assert.Equal(t, argoCD, NewArgoCD(argoCDReg.URL, argoCDReg.Token, argoCDReg.Namespace))

	argoCD, err = factory.GetArgoCD("not-exists")
	assert.Nil(t, argoCD)
	assert.NotNil(t, err)
}
