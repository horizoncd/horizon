package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _   = orm.NewSqliteDB("")
	ctx     = orm.NewContext(context.TODO(), db)
	manager = managerparam.InitManager(db)
)

// nolint
func init() {
	// create table
	err := db.AutoMigrate(&clustermodels.Cluster{}, &applicationmodels.Application{}, &groupmodels.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}

func TestServiceGetByID(t *testing.T) {
	group := &groupmodels.Group{
		Name:         "a",
		Path:         "a",
		TraversalIDs: "1",
	}
	db.Save(group)

	application := &applicationmodels.Application{
		Name:    "b",
		GroupID: group.ID,
	}
	db.Save(application)

	cluster := &clustermodels.Cluster{
		Name:          "c",
		ApplicationID: application.ID,
	}
	db.Save(cluster)

	t.Run("GetByID", func(t *testing.T) {
		s := service{
			applicationService: applicationservice.NewService(groupservice.NewService(manager), manager),
			clusterManager:     manager.ClusterMgr,
		}
		result, err := s.GetByID(ctx, application.ID)
		assert.Nil(t, err)
		assert.Equal(t, "/a/b/c", result.FullPath)
	})
}
