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
	"os"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	appregionmanager "github.com/horizoncd/horizon/pkg/applicationregion/manager"
	appregionmodels "github.com/horizoncd/horizon/pkg/applicationregion/models"
	"github.com/horizoncd/horizon/pkg/environment/models"
	envregionmanager "github.com/horizoncd/horizon/pkg/environmentregion/manager"
	envregion "github.com/horizoncd/horizon/pkg/environmentregion/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	"github.com/stretchr/testify/assert"
)

var (
	db, _        = orm.NewSqliteDB("")
	regionMgr    = regionmanager.New(db)
	ctx          context.Context
	mgr          = New(db)
	appregionMgr = appregionmanager.New(db)
	envregionMgr = envregionmanager.New(db)
)

func Test(t *testing.T) {
	_, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	_, err = regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz-update",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	onlineEnv, err := mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "online",
		DisplayName: "线上",
	})
	assert.Nil(t, err)
	t.Logf("%v", onlineEnv)
	err = mgr.UpdateByID(ctx, onlineEnv.ID, &models.Environment{
		Name:        "online-update",
		DisplayName: "线上-update",
	})
	assert.Nil(t, err)
	env, err := mgr.GetByID(ctx, onlineEnv.ID)
	assert.Nil(t, err)
	assert.Equal(t, env.DisplayName, "线上-update")

	env, err = mgr.GetByName(ctx, onlineEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, env.DisplayName, "线上-update")

	preEnv, err := mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "pre",
		DisplayName: "预发",
	})
	assert.Nil(t, err)
	t.Logf("%v", preEnv)

	testEnv, err := mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "test",
		DisplayName: "测试",
	})
	assert.Nil(t, err)
	t.Logf("%v", testEnv)

	devEnv, err := mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "dev",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	t.Logf("%v", devEnv)

	envs, err := mgr.ListAllEnvironment(ctx)
	assert.Nil(t, err)
	assert.Equal(t, len(envs), 4)
	t.Logf("%v", envs[0])
	t.Logf("%v", envs[1])
	t.Logf("%v", envs[2])
	t.Logf("%v", envs[3])

	err = appregionMgr.UpsertByApplicationID(ctx, uint(1), []*appregionmodels.ApplicationRegion{
		{
			ID:              0,
			ApplicationID:   uint(1),
			EnvironmentName: "dev",
			RegionName:      "",
		},
	})
	assert.Nil(t, err)
	_, err = envregionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "dev",
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	regionParts, err := envregionMgr.ListEnabledRegionsByEnvironment(ctx, "dev")
	assert.Nil(t, err)
	assert.Equal(t, len(regionParts), 1)
	assert.Equal(t, regionParts[0].Name, "hz")
	assert.Equal(t, regionParts[0].DisplayName, "HZ")

	err = mgr.DeleteByID(ctx, devEnv.ID)
	assert.Nil(t, err)

	applicationRegions, _ := appregionMgr.ListByApplicationID(ctx, uint(1))
	assert.Empty(t, applicationRegions)

	regions, _ := envregionMgr.ListByEnvironment(ctx, devEnv.Name)
	assert.Empty(t, regions)
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.Environment{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&envregion.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&appregionmodels.ApplicationRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&envregionmodels.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}
