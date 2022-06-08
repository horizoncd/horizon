package handler

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/application"
	"g.hz.netease.com/horizon/core/controller/cluster"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/cmdb"
	cmdbconfig "g.hz.netease.com/horizon/pkg/config/cmdb"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"github.com/stretchr/testify/assert"
)

var cmdbctl cmdb.Controller
var handler EventHandler
var applicationName = "horizon-tmo-test"
var clusterName = "horizon-tmo-test-1"

// nolint
func TestApplication(t *testing.T) {
	ret := &application.GetApplicationResponse{
		CreateApplicationRequest: application.CreateApplicationRequest{
			Base: application.Base{
				Priority: "P2",
			},
			Name: applicationName,
		},
	}
	var createUser userauth.User = &userauth.DefaultInfo{
		Name:     "hzsunjianliang",
		FullName: "cat",
	}
	ctx := context.WithValue(context.TODO(), common.UserContextKey(), createUser)

	// 1. create application
	createApplicationEvent := &hook.EventCtx{
		EventType: hook.CreateApplication,
		Event:     ret,
		Ctx:       ctx,
	}
	assert.Nil(t, handler.Process(createApplicationEvent))

	// 2. create cluster
	ret2 := &cluster.GetClusterResponse{
		CreateClusterRequest: &cluster.CreateClusterRequest{
			Name: clusterName,
		},
		Application: &cluster.Application{
			Name: applicationName,
		},
		Scope: &cluster.Scope{
			Environment: "test",
		},
	}
	createClusterEvent := &hook.EventCtx{
		EventType: hook.CreateCluster,
		Event:     ret2,
		Ctx:       ctx,
	}
	assert.Nil(t, handler.Process(createClusterEvent))

	// 3. delete cluster
	deleteClusterEvent := &hook.EventCtx{
		EventType: hook.DeleteCluster,
		Event:     clusterName,
		Ctx:       ctx,
	}
	assert.Nil(t, handler.Process(deleteClusterEvent))

	// 4. delete application
	deleteApplicationEvent := &hook.EventCtx{
		EventType: hook.DeleteApplication,
		Event:     applicationName,
		Ctx:       ctx,
	}
	assert.Nil(t, handler.Process(deleteApplicationEvent))
}

func TestMain(m *testing.M) {
	config := cmdbconfig.Config{
		URL:        "api-in.nss.netease.com",
		ClientID:   "musicHorizon",
		SecretCode: "",
		ParentID:   10,
	}
	cmdbctl = cmdb.NewController(config)
	handler = NewCMDBEventHandler(cmdbctl)
	os.Exit(m.Run())
}
