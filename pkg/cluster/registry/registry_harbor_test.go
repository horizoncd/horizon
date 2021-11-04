package registry

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"g.hz.netease.com/horizon/pkg/cluster/registry/mock"
)

var harbor = &HarborConfig{}
var server = mock.NewHarborServer()

func TestMain(m *testing.M) {
	s := httptest.NewServer(http.HandlerFunc(server.R.ServeHTTP))
	harbor.Server = "http://" + s.Listener.Addr().String()
	os.Exit(m.Run())
}

func TestByMock(t *testing.T) {
	h := NewHarborRegistry(harbor)
	ctx := context.Background()
	// add project1
	projectID, err := h.CreateProject(ctx, "project1")
	assert.Nil(t, err)
	fmt.Printf("projectID: %d", projectID)
	// add project1 again
	projectIDAgain, err := h.CreateProject(ctx, "project1")
	assert.Nil(t, err)
	assert.Equal(t, -1, projectIDAgain)

	// 推送镜像到repo1
	server.PushImage("project1", "repo1", "v1")
	// 删除repo1
	err = h.DeleteRepository(ctx, "project1", "repo1")
	assert.Nil(t, err)
	// 再次删除repo1
	err = h.DeleteRepository(ctx, "project1", "repo1")
	assert.Nil(t, err)

	// 推送镜像到repo2
	server.PushImage("project1", "repo2", "master-12345578-20210702134536")
	server.PushImage("project1", "repo2", "master-12345578-20210703113624")
	server.PushImage("project1", "repo2", "master-12345578-20210704100908")
	server.PushImage("project1", "repo2", "master-12345578-20210703100908")
	images, err := h.ListImage(ctx, "project1", "repo2")
	assert.Nil(t, err)
	wantImages := []string{
		strings.TrimPrefix(strings.TrimPrefix(harbor.Server, "http://"), "https://") +
			"/project1/repo2:master-12345578-20210704100908",
		strings.TrimPrefix(strings.TrimPrefix(harbor.Server, "http://"), "https://") +
			"/project1/repo2:master-12345578-20210703113624",
		strings.TrimPrefix(strings.TrimPrefix(harbor.Server, "http://"), "https://") +
			"/project1/repo2:master-12345578-20210703100908",
		strings.TrimPrefix(strings.TrimPrefix(harbor.Server, "http://"), "https://") +
			"/project1/repo2:master-12345578-20210702134536",
	}
	assert.Equal(t, wantImages, images)

	// project3 is not existed
	_, err = h.ListImage(ctx, "project3", "repo2")
	assert.NotNil(t, err)

	s := h.GetServer(ctx)
	assert.Equal(t, harbor.Server, s)
}
