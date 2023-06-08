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

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	perror "github.com/horizoncd/horizon/pkg/errors"
	envmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/stretchr/testify/assert"
)

func createEnvironmentRegionCtx() (context.Context, RegionManager, EnvironmentManager, EnvironmentRegionManager) {
	var (
		db, _     = orm.NewSqliteDB("")
		ctx       context.Context
		regionMgr = NewRegionManager(db)
		envMgr    = NewEnvironmentManager(db)
		mgr       = NewEnvironmentRegionManager(db)
	)

	if err := db.AutoMigrate(&envmodels.Environment{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&envmodels.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&envmodels.Region{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	return ctx, regionMgr, envMgr, mgr
}

func TestEnvironmentRegion(t *testing.T) {
	ctx, regionMgr, envMgr, mgr := createEnvironmentRegionCtx()
	_, err := regionMgr.Create(ctx, &envmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	onlineEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "online",
		DisplayName: "线上",
	})
	assert.Nil(t, err)
	t.Logf("%v", onlineEnv)

	preEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "pre",
		DisplayName: "预发",
	})
	assert.Nil(t, err)
	t.Logf("%v", preEnv)

	testEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "test",
		DisplayName: "测试",
	})
	assert.Nil(t, err)
	t.Logf("%v", testEnv)

	devEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "dev",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	t.Logf("%v", devEnv)

	devHzEr, err := mgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
		AutoFree:        true,
		IsDefault:       true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, devHzEr)
	t.Logf("%v", devHzEr)

	_, err = mgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
	})
	assert.NotNil(t, err)
	t.Logf("%v", err)

	regions, err := mgr.ListByEnvironment(ctx, devEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, "hz", regions[0].RegionName)
	assert.Equal(t, true, regions[0].AutoFree)
	t.Logf("%v", regions[0])

	err = mgr.SetEnvironmentRegionIfAutoFree(ctx, devHzEr.ID, false)
	assert.Nil(t, err)

	regions, err = mgr.ListByEnvironment(ctx, devEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, false, regions[0].AutoFree)

	// test SetEnvironmentRegionToDefaultByID
	r2, _ := mgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz1",
	})

	err = mgr.SetEnvironmentRegionToDefaultByID(ctx, r2.ID)
	assert.Nil(t, err)
	r2New, err := mgr.GetEnvironmentRegionByID(ctx, r2.ID)
	assert.Nil(t, err)
	assert.Equal(t, r2New.IsDefault, true)
	r2New, err = mgr.GetDefaultRegionByEnvironment(ctx, devEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, r2New.IsDefault, true)

	devHzErNew, err := mgr.GetEnvironmentRegionByID(ctx, devHzEr.ID)
	assert.Nil(t, err)
	assert.Equal(t, devHzErNew.IsDefault, false)
	devHzErNew, err = mgr.GetByEnvironmentAndRegion(ctx, devEnv.Name, "hz")
	assert.Nil(t, err)
	assert.Equal(t, devHzErNew.IsDefault, false)

	// test deleteByID
	err = mgr.DeleteByID(ctx, devHzErNew.ID)
	assert.Nil(t, err)
	_, err = mgr.GetEnvironmentRegionByID(ctx, devHzEr.ID)
	_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)
}
