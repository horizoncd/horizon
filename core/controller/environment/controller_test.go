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

package environment

import (
	"context"
	"os"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/region"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/service"
	"github.com/stretchr/testify/assert"
)

var (
	ctx     context.Context
	manager *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&models.Environment{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&models.Region{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		ID: uint(1),
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	param := &param.Param{
		AutoFreeSvc: service.NewAutoFreeSVC([]string{}),
		Manager:     manager,
	}
	regionCtl := region.NewController(param)
	_, err := regionCtl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	_, err = regionCtl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz-update",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)
	ctl := NewController(param)
	envs, err := ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(envs))

	devID, err := ctl.Create(ctx, &CreateEnvironmentRequest{
		Name:        "dev",
		DisplayName: "DEV",
	})
	assert.Nil(t, err)
	env, err := ctl.GetByID(ctx, devID)
	assert.Nil(t, err)
	assert.Equal(t, "dev", env.Name)
	assert.Equal(t, false, env.AutoFree)

	envByName, err := ctl.GetByName(ctx, env.Name)
	assert.Nil(t, err)
	assert.Equal(t, env, envByName)

	envs, err = ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "dev", envs[0].Name)
	assert.Equal(t, "DEV", envs[0].DisplayName)
	assert.Equal(t, false, envs[0].AutoFree)

	err = ctl.UpdateByID(ctx, devID, &UpdateEnvironmentRequest{
		DisplayName: "DEV-update",
	})
	assert.Nil(t, err)

	envs, err = ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "dev", envs[0].Name)
	assert.Equal(t, "DEV-update", envs[0].DisplayName)
}
