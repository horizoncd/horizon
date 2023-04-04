// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package service

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// nolint
func createClusterCtx() (context.Context, *managerparam.Manager, *gorm.DB) {
	var (
		// use tmp sqlite
		db, _   = orm.NewSqliteDB("")
		ctx     = context.TODO()
		manager = managerparam.InitManager(db)
	)
	// create table
	err := db.AutoMigrate(&models.Cluster{}, &models.Application{}, &models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	return ctx, manager, db
}

func TestClusterServiceGetByID(t *testing.T) {
	ctx, manager, db := createClusterCtx()
	group := &models.Group{
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

	cluster := &models.Cluster{
		Name:          "c",
		ApplicationID: application.ID,
	}
	db.Save(cluster)

	t.Run("GetByID", func(t *testing.T) {
		s := clusterService{
			applicationService: NewApplicationService(NewGroupService(manager), manager),
			clusterManager:     manager.ClusterMgr,
		}
		result, err := s.GetByID(ctx, application.ID)
		assert.Nil(t, err)
		assert.Equal(t, "/a/b/c", result.FullPath)
	})
}
