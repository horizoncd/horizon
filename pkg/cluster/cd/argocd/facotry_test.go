package argocd

import (
	"testing"

	"g.hz.netease.com/horizon/pkg/config/argocd"
	"github.com/stretchr/testify/assert"
)

func Test(t *testing.T) {
	argoCDMapper := make(map[string]*argocd.ArgoCD)
	argoCDTest := &argocd.ArgoCD{
		URL:   "http://test.argo.com",
		Token: "token1",
	}
	argoCDReg := &argocd.ArgoCD{
		URL:   "http://reg.argo.com",
		Token: "token1",
	}
	argoCDMapper["test"] = argoCDTest
	argoCDMapper["reg"] = argoCDReg
	factory := NewFactory(argoCDMapper)
	assert.NotNil(t, factory)

	argoCD, err := factory.GetArgoCD("test")
	assert.Nil(t, err)
	assert.NotNil(t, argoCD)
	assert.Equal(t, argoCD, NewArgoCD(argoCDTest.URL, argoCDTest.Token))

	argoCD, err = factory.GetArgoCD("reg")
	assert.Nil(t, err)
	assert.NotNil(t, argoCD)
	assert.Equal(t, argoCD, NewArgoCD(argoCDReg.URL, argoCDReg.Token))

	argoCD, err = factory.GetArgoCD("not-exists")
	assert.Nil(t, argoCD)
	assert.NotNil(t, err)
}
