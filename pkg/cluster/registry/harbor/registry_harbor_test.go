package harbor

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"g.hz.netease.com/horizon/pkg/cluster/registry"
	"g.hz.netease.com/horizon/pkg/cluster/registry/mock"
	"github.com/stretchr/testify/assert"
)

var harbor = &registry.Config{}
var server = mock.NewHarborServer()

func TestMain(m *testing.M) {
	s := httptest.NewServer(http.HandlerFunc(server.R.ServeHTTP))
	harbor.Server = "http://" + s.Listener.Addr().String()
	os.Exit(m.Run())
}

func TestByMock(t *testing.T) {
	harbor.Path = "project1"
	registry, _ := NewHarborRegistry(harbor)
	h := registry.(*Registry)
	ctx := context.Background()
	// add project1
	projectID, err := h.createProject(ctx, "project1")
	assert.Nil(t, err)
	fmt.Printf("projectID: %d", projectID)
	// add project1 again
	projectIDAgain, err := h.createProject(ctx, "project1")
	assert.Nil(t, err)
	assert.Equal(t, -1, projectIDAgain)

	// 推送镜像到repo1
	server.PushImage("project1", "repo1", "v1")
	// 删除repo1
	err = h.DeleteRepository(ctx, "repo1")
	assert.Nil(t, err)
	// 再次删除repo1
	err = h.DeleteRepository(ctx, "repo1")
	assert.Nil(t, err)

	// 推送镜像到repo2
	server.PushImage("project1", "repo2", "master-12345578-20210702134536")
	server.PushImage("project1", "repo2", "master-12345578-20210703113624")
	server.PushImage("project1", "repo2", "master-12345578-20210704100908")
	server.PushImage("project1", "repo2", "master-12345578-20210703100908")
}
