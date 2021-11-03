package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupModels "g.hz.netease.com/horizon/pkg/group/models"
	groupservice "g.hz.netease.com/horizon/pkg/group/service"
	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   = orm.NewContext(context.TODO(), db)
)

// nolint
func init() {
	// create table
	err := db.AutoMigrate(&models.Application{}, &groupModels.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}

func TestServiceGetByID(t *testing.T) {
	group := &groupModels.Group{
		Name:         "a",
		Path:         "a",
		TraversalIDs: "1",
	}
	db.Save(group)

	application := &models.Application{
		Name:    "b",
		GroupID: group.ID,
	}
	db.Save(application)

	t.Run("GetByID", func(t *testing.T) {
		s := service{
			groupService:       groupservice.Svc,
			applicationManager: manager.Mgr,
		}
		result, err := s.GetByID(ctx, application.ID)
		assert.Nil(t, err)
		assert.Equal(t, "/a/b", result.FullPath)
	})
}
