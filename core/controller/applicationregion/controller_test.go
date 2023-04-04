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

package applicationregion

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	appmodels "github.com/horizoncd/horizon/pkg/models"
	tagmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/stretchr/testify/assert"
)

var (
	ctx context.Context
	c   Controller
	mgr *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	mgr = managerparam.InitManager(db)
	if err := db.AutoMigrate(&appmodels.ApplicationRegion{}, &appmodels.Region{},
		&appmodels.Environment{}, &appmodels.EnvironmentRegion{}, &appmodels.Tag{},
		appmodels.Group{}, appmodels.Application{}, appmodels.Member{},
	); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   1,
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	regionMgr := mgr.RegionMgr
	r1, err := regionMgr.Create(ctx, &appmodels.Region{
		Name:        "hz-test",
		DisplayName: "杭州测试",
	})
	assert.Nil(t, err)
	assert.NotNil(t, r1)
	r2, err := regionMgr.Create(ctx, &appmodels.Region{
		Name:        "hz",
		DisplayName: "杭州",
	})
	assert.Nil(t, err)
	assert.NotNil(t, r2)

	r3, err := regionMgr.Create(ctx, &appmodels.Region{
		Name:        "singapore",
		DisplayName: "新加坡",
	})
	assert.Nil(t, err)
	assert.NotNil(t, r3)
	err = mgr.TagManager.UpsertByResourceTypeID(ctx, common.ResourceRegion, r1.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
	})
	assert.Nil(t, err)
	err = mgr.TagManager.UpsertByResourceTypeID(ctx, common.ResourceRegion, r2.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
	})
	assert.Nil(t, err)
	err = mgr.TagManager.UpsertByResourceTypeID(ctx, common.ResourceRegion, r3.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
	})
	assert.Nil(t, err)

	envMgr := mgr.EnvMgr
	envs := make([]*appmodels.Environment, 0)
	for _, env := range []string{"test", "beta", "perf", "pre", "online"} {
		environment, err := envMgr.CreateEnvironment(ctx, &appmodels.Environment{
			Name: env,
		})
		assert.Nil(t, err)
		envs = append(envs, environment)
	}

	_, err = mgr.EnvRegionMgr.CreateEnvironmentRegion(ctx, &appmodels.EnvironmentRegion{
		EnvironmentName: envs[0].Name,
		RegionName:      r1.Name,
	})
	assert.Nil(t, err)

	_, err = mgr.EnvRegionMgr.CreateEnvironmentRegion(ctx, &appmodels.EnvironmentRegion{
		EnvironmentName: envs[3].Name,
		RegionName:      r2.Name,
		IsDefault:       true,
	})
	assert.Nil(t, err)

	_, err = mgr.EnvRegionMgr.CreateEnvironmentRegion(ctx, &appmodels.EnvironmentRegion{
		EnvironmentName: envs[3].Name,
		RegionName:      r3.Name,
	})
	assert.Nil(t, err)

	c = &controller{
		mgr:                  mgr.ApplicationRegionManager,
		regionMgr:            regionMgr,
		environmentMgr:       envMgr,
		environmentRegionMgr: mgr.EnvironmentRegionMgr,
		applicationMgr:       mgr.ApplicationManager,
		groupMgr:             mgr.GroupManager,
	}

	g, err := mgr.GroupManager.Create(ctx, &appmodels.Group{
		Name: "",
		Path: "",
		RegionSelector: `- key: "a"
  values: 
    - "1"`,
	})
	assert.Nil(t, err)
	application, err := mgr.ApplicationManager.Create(ctx, &appmodels.Application{
		Model:   global.Model{},
		GroupID: g.ID,
		Name:    "",
	}, map[string]string{})
	assert.Nil(t, err)

	applicationID := application.ID

	regions, err := c.List(ctx, applicationID)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(regions))
	assert.Equal(t, "hz", getRegionByEnvironment("pre", regions))
	b, _ := json.Marshal(regions)
	t.Logf("%v", string(b))

	applicationRegions := []*Region{
		{
			Environment: "test",
			Region:      "hz-test",
		},
		{
			Environment: "pre",
			Region:      "hz",
		},
	}
	err = c.Update(ctx, applicationID, applicationRegions)
	assert.Nil(t, err)

	applicationRegions2 := []*Region{
		{
			Environment: "pre",
			Region:      "hz-test",
		},
		{
			Environment: "test",
			Region:      "hz",
		},
	}
	err = c.Update(ctx, applicationID, applicationRegions2)
	assert.NotNil(t, err)
	t.Logf("%v", err)

	regions, err = c.List(ctx, applicationID)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(regions))
	assert.Equal(t, "hz", getRegionByEnvironment("pre", regions))
	b, _ = json.Marshal(regions)
	t.Logf("%v", string(b))

	applicationRegions = []*Region{
		{
			Environment: "pre",
			Region:      "singapore",
		},
	}
	err = c.Update(ctx, applicationID, applicationRegions)
	assert.Nil(t, err)
	regions, err = c.List(ctx, applicationID)
	assert.Nil(t, err)
	assert.Equal(t, 5, len(regions))
	assert.Equal(t, "singapore", getRegionByEnvironment("pre", regions))
	b, _ = json.Marshal(regions)
	t.Logf("%v", string(b))
}

func getRegionByEnvironment(environment string, regions []*Region) string {
	for _, r := range regions {
		if r.Environment == environment {
			return r.Region
		}
	}
	return ""
}
