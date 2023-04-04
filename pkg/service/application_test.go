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
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createApplicationCtx() (context.Context, *managerparam.Manager, *gorm.DB) {
	var (
		db, _ = orm.NewSqliteDB("")
		ctx   = context.TODO()
		mgr   = managerparam.InitManager(db)
	)

	err := db.AutoMigrate(&models.Application{}, &models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		panic(err)
	}
	return ctx, mgr, db
}

func TestApplicationServiceGetByID(t *testing.T) {
	ctx, mgr, db := createApplicationCtx()
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

	t.Run("GetByID", func(t *testing.T) {
		s := applicationService{
			groupSvc: NewGroupService(mgr),
			appMgr:   mgr.ApplicationManager,
		}
		result, err := s.GetByID(ctx, application.ID)
		assert.Nil(t, err)
		assert.Equal(t, "/a/b", result.FullPath)
	})

	t.Run("GetByIDs", func(t *testing.T) {
		s := applicationService{
			groupSvc: NewGroupService(mgr),
			appMgr:   mgr.ApplicationManager,
		}
		result, err := s.GetByIDs(ctx, []uint{application.ID})
		assert.Nil(t, err)
		assert.Equal(t, 1, len(result))
		assert.Equal(t, "/a/b", result[application.ID].FullPath)
	})
}
