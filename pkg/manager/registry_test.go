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

package manager

import (
	"context"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	regionmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createRegistryCtx() (context.Context, RegistryManager, *gorm.DB) {
	var (
		// use tmp sqlite
		db, _ = orm.NewSqliteDB("")
		ctx   context.Context
		mgr   = NewRegistryManager(db)
	)
	if err := db.AutoMigrate(&regionmodels.Registry{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}

	ctx = context.TODO()

	return ctx, mgr, db
}

func TestRegistry(t *testing.T) {
	ctx, mgr, _ := createRegistryCtx()
	id, err := mgr.Create(ctx, &regionmodels.Registry{
		Name:   "1",
		Server: "2",
		Token:  "1",
	})
	assert.Nil(t, err)

	registry, err := mgr.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, registry.Name, "1")
	assert.Equal(t, registry.Server, "2")
	assert.Equal(t, registry.Token, "1")

	err = mgr.UpdateByID(ctx, id, &regionmodels.Registry{
		Name:   "2",
		Server: "1",
		Token:  "2",
	})
	assert.Nil(t, err)
	registry, _ = mgr.GetByID(ctx, id)
	assert.Equal(t, registry.Name, "2")
	assert.Equal(t, registry.Server, "1")
	assert.Equal(t, registry.Token, "2")

	err = mgr.DeleteByID(ctx, id)
	assert.Nil(t, err)
	registry, err = mgr.GetByID(ctx, id)
	assert.NotNil(t, err)
	assert.Nil(t, registry)
}
