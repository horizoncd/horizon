package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	applicationmodels "github.com/horizoncd/horizon/pkg/application/models"
	applicationservice "github.com/horizoncd/horizon/pkg/application/service"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	groupservice "github.com/horizoncd/horizon/pkg/group/service"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _   = orm.NewSqliteDB("")
	ctx     = context.TODO()
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
