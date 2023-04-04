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
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createUserCtx() (context.Context, *managerparam.Manager, *gorm.DB) {
	var (
		db      *gorm.DB
		ctx     context.Context
		manager *managerparam.Manager
	)

	db, _ = orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&models.User{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	return ctx, manager, db
}

func Test_service_CheckUsersExists(t *testing.T) {
	ctx, manager, _ := createUserCtx()
	userManger := manager.UserManager
	_, err := userManger.Create(ctx, &models.User{
		Name:  "tony",
		Email: "tony@corp.com",
	})
	assert.Nil(t, err)

	_, err = userManger.Create(ctx, &models.User{
		Name:  "mary",
		Email: "mary@corp.com",
	})
	assert.Nil(t, err)

	svc := NewUserService(manager)
	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com"})
	assert.Nil(t, err)

	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com", "mary@corp.com"})
	assert.Nil(t, err)

	err = svc.CheckUsersExists(ctx, []string{"tony@corp.com", "not-exist@corp.com"})
	assert.NotNil(t, err)
	t.Logf("%v", err)
}
