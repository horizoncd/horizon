package cmdb

import (
	"context"
	"testing"

	"g.hz.netease.com/horizon/pkg/config/cmdb"
	"github.com/stretchr/testify/assert"
)

var applicationName string = "horizon-test"
var clusterName string = "horizon-test-cluster-1"

func TestCreateApplication(t *testing.T) {
	config := cmdb.Config{
		URL:        "api.nss.netease.com",
		ClientID:   "musicHorizon",
		SecretCode: "",
	}
	c := NewController(config)

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
}
