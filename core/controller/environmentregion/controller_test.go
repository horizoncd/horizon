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

package environmentregion

import (
	"context"
	"os"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/environment"
	"github.com/horizoncd/horizon/core/controller/region"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	envmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
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
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		ID: uint(1),
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	regionCtl := region.NewController(&param.Param{
		Manager: manager,
	})
	_, err := regionCtl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	environmentCtl := environment.NewController(&param.Param{Manager: manager})
	_, err = environmentCtl.Create(ctx, &environment.CreateEnvironmentRequest{
		Name:        "dev",
		DisplayName: "DEV",
	})
	assert.Nil(t, err)

	ctl := NewController(&param.Param{Manager: manager})
	er, err := ctl.CreateEnvironmentRegion(ctx, &CreateEnvironmentRegionRequest{
		EnvironmentName: "dev",
		RegionName:      "hz",
		AutoFree:        true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)

	ers, err := ctl.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ers))
	assert.Equal(t, "hz", ers[0].RegionName)
	assert.Equal(t, true, ers[0].AutoFree)

	err = ctl.SetEnvironmentRegionIfAutoFree(ctx, er, false)
	assert.Nil(t, err)

	ers, err = ctl.ListByEnvironment(ctx, "dev")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ers))
	assert.Equal(t, "hz", ers[0].RegionName)
	assert.Equal(t, false, ers[0].AutoFree)

	err = ctl.SetEnvironmentRegionToDefault(ctx, er)
	assert.Nil(t, err)

	ers, err = ctl.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ers))
	assert.Equal(t, true, ers[0].IsDefault)

	err = ctl.DeleteByID(ctx, er)
	assert.Nil(t, err)

	ers, err = ctl.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ers))
}
