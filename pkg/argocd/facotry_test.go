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
	// env-argocd mapper
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

	// region-argocd mapper
	regionArgoCDMapper := make(map[string]*argocd.ArgoCD)
	argoCDTest2 := &argocd.ArgoCD{
		URL:       "http://test2.argo.com",
		Token:     "token2",
		Namespace: "argocd2",
	}
	regionArgoCDMapper["region"] = argoCDTest2

	// create factory
	factory := NewFactory(argoCDMapper, regionArgoCDMapper)
	assert.NotNil(t, factory)

	// 1. use test argocd from region-argocd mapper
	argoCD, err := factory.GetArgoCD("region", "test")
	assert.Nil(t, err)
	assert.NotNil(t, argoCD)
	assert.Equal(t, NewArgoCD(argoCDTest2.URL, argoCDTest2.Token, argoCDTest2.Namespace), argoCD)

	// 2. use reg argocd from env-argocd mapper
	argoCD, err = factory.GetArgoCD("region2", "reg")
	assert.Nil(t, err)
	assert.NotNil(t, argoCD)
	assert.Equal(t, NewArgoCD(argoCDReg.URL, argoCDReg.Token, argoCDReg.Namespace), argoCD)

	// 3. use test argocd from env-argocd mapper
	argoCD, err = factory.GetArgoCD("region2", "test")
	assert.Nil(t, err)
	assert.NotNil(t, argoCD)
	assert.Equal(t, NewArgoCD(argoCDTest.URL, argoCDTest.Token, argoCDTest.Namespace), argoCD)

	// 4. report error because not found in region-argocd mapper and env-argocd mapper
	argoCD, err = factory.GetArgoCD("region2", "not-exists")
	assert.Nil(t, argoCD)
	assert.NotNil(t, err)
}
