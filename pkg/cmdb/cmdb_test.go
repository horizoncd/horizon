package cmdb

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/pkg/config/cmdb"
	"github.com/stretchr/testify/assert"
)

var applicationName string = "horizon-test"
var clusterName string = "horizon-test-cluster-1"

var c Controller

func TestCreateApplication(t *testing.T) {
	admins := make([]Account, 0)
	admins = append(admins, Account{
		Account:     "hzsunjianliang",
		AccountType: "user",
	})
	req := CreateApplicationRequest{
		Name:     applicationName,
		ParentID: 10,
		Priority: P2,
		Admin:    admins,
	}

	assert.Nil(t, c.CreateApplication(context.TODO(), req))
	assert.Nil(t, c.DeleteApplication(context.TODO(), applicationName))
}

func TestCreateCluster(t *testing.T) {
	ctx := context.TODO()
	admins := make([]Account, 0)
	admins = append(admins, Account{
		Account:     "hzsunjianliang",
		AccountType: "user",
	})
	createAppReq := CreateApplicationRequest{
		Name:     applicationName,
		ParentID: 10,
		Priority: P2,
		Admin:    admins,
	}
	assert.Nil(t, c.CreateApplication(ctx, createAppReq))

	createClusterReq := CreateClusterRequest{
		Name:                clusterName,
		ApplicationName:     applicationName,
		Env:                 Test,
		ClusterServerStatus: StatusReady,
		ClusterStyle:        "docker",
		Admin:               admins,
	}
	assert.Nil(t, c.CreateCluster(ctx, createClusterReq))
	assert.Nil(t, c.DeleteCluster(ctx, clusterName))
	assert.Nil(t, c.DeleteCluster(ctx, clusterName))
}

func TestMain(m *testing.M) {
	config := cmdb.Config{
		URL:        "api-in.nss.netease.com",
		ClientID:   "musicHorizon",
		SecretCode: "",
		ParentID:   10,
	}
	c = NewController(config)
	os.Exit(m.Run())
}
