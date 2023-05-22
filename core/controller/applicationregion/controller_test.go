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
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/applicationregion/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	envmodels "github.com/horizoncd/horizon/pkg/environment/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/stretchr/testify/assert"
)

var (
	ctx     context.Context
	c       Controller
	manager *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&models.ApplicationRegion{}, &regionmodels.Region{},
		&envmodels.Environment{}, &envregionmodels.EnvironmentRegion{}, &tagmodels.Tag{},
		groupmodels.Group{}, appmodels.Application{}, membermodels.Member{},
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
	regionMgr := manager.RegionMgr
	r1, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz-test",
		DisplayName: "杭州测试",
	})
	assert.Nil(t, err)
	assert.NotNil(t, r1)
	r2, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "杭州",
	})
	assert.Nil(t, err)
	assert.NotNil(t, r2)

	r3, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "singapore",
		DisplayName: "新加坡",
	})
	assert.Nil(t, err)
	assert.NotNil(t, r3)
	err = manager.TagManager.UpsertByResourceTypeID(ctx, common.ResourceRegion, r1.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
	})
	assert.Nil(t, err)
	err = manager.TagManager.UpsertByResourceTypeID(ctx, common.ResourceRegion, r2.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
	})
	assert.Nil(t, err)
	err = manager.TagManager.UpsertByResourceTypeID(ctx, common.ResourceRegion, r3.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
	})
	assert.Nil(t, err)

	envMgr := manager.EnvMgr
	envs := make([]*envmodels.Environment, 0)
	for _, env := range []string{"test", "beta", "perf", "pre", "online"} {
		environment, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
			Name: env,
		})
		assert.Nil(t, err)
		envs = append(envs, environment)
	}

	_, err = manager.EnvRegionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: envs[0].Name,
		RegionName:      r1.Name,
	})
	assert.Nil(t, err)

	_, err = manager.EnvRegionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: envs[3].Name,
		RegionName:      r2.Name,
		IsDefault:       true,
	})
	assert.Nil(t, err)

	_, err = manager.EnvRegionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: envs[3].Name,
		RegionName:      r3.Name,
	})
	assert.Nil(t, err)

	c = &controller{
		mgr:                  manager.ApplicationRegionManager,
		regionMgr:            regionMgr,
		environmentMgr:       envMgr,
		environmentRegionMgr: manager.EnvironmentRegionMgr,
		applicationMgr:       manager.ApplicationManager,
		groupMgr:             manager.GroupManager,
	}

	g, err := manager.GroupManager.Create(ctx, &groupmodels.Group{
		Name: "",
		Path: "",
		RegionSelector: `- key: "a"
  values: 
    - "1"`,
	})
	assert.Nil(t, err)
	application, err := manager.ApplicationManager.Create(ctx, &appmodels.Application{
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
