package v1

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/horizoncd/horizon/pkg/cluster/registry"
	"github.com/horizoncd/horizon/pkg/cluster/registry/harbor/v1/mockserver"
	"github.com/stretchr/testify/assert"
)

var config = &registry.Config{}
var server = mockserver.NewHarborServer()

func TestMain(m *testing.M) {
	s := httptest.NewServer(http.HandlerFunc(server.R.ServeHTTP))
	config.Server = "http://" + s.Listener.Addr().String()
	os.Exit(m.Run())
}

func TestByMock(t *testing.T) {
	config.Path = "project1"
	registry, _ := NewHarborRegistry(config)
	h := registry.(*Registry)
	ctx := context.Background()

	server.CreateProject("project1", nil)
	server.PushImage("project1", "horizon-demo/horizon-demo-dev", "v1")

	err := h.DeleteImage(ctx, "horizon-demo", "horizon-demo-dev")
	assert.Nil(t, err)
	err = h.DeleteImage(ctx, "horizon-demo", "horizon-demo-dev")
	assert.Nil(t, err)
}
