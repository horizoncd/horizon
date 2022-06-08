package applicationregion

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/applicationregion/manager"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	envregionmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	envreigonmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/server/global"
	tagmanager "g.hz.netease.com/horizon/pkg/tag/manager"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	"github.com/stretchr/testify/assert"
)

var (
	ctx context.Context
	c   Controller
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.ApplicationRegion{}, &regionmodels.Region{},
		&envmodels.Environment{}, &envregionmodels.EnvironmentRegion{}, &tagmodels.Tag{},
		groupmodels.Group{}, appmodels.Application{}, membermodels.Member{},
	); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   1,
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	regionMgr := regionmanager.Mgr
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
	err = tagmanager.Mgr.UpsertByResourceTypeID(ctx, tagmodels.TypeRegion, r1.ID, []*tagmodels.Tag{
		{
			ResourceID:   r1.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "a",
			Value:        "1",
		},
	})
	assert.Nil(t, err)
	err = tagmanager.Mgr.UpsertByResourceTypeID(ctx, tagmodels.TypeRegion, r1.ID, []*tagmodels.Tag{
		{
			ResourceID:   r2.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "a",
			Value:        "1",
		},
	})
	assert.Nil(t, err)
	err = tagmanager.Mgr.UpsertByResourceTypeID(ctx, tagmodels.TypeRegion, r1.ID, []*tagmodels.Tag{
		{
			ResourceID:   r3.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "a",
			Value:        "1",
		},
	})
	assert.Nil(t, err)

	envMgr := envmanager.Mgr
	envs := make([]*envmodels.Environment, 0)
	for _, env := range []string{"test", "beta", "perf", "pre", "online"} {
		environment, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
			Name: env,
		})
		assert.Nil(t, err)
		envs = append(envs, environment)
	}

	_, err = envreigonmanager.Mgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: envs[0].Name,
		RegionName:      r1.Name,
	})
	assert.Nil(t, err)

	_, err = envreigonmanager.Mgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: envs[3].Name,
		RegionName:      r2.Name,
		IsDefault:       true,
	})
	assert.Nil(t, err)

	_, err = envreigonmanager.Mgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: envs[3].Name,
		RegionName:      r3.Name,
	})
	assert.Nil(t, err)

	c = &controller{
		mgr:                  manager.Mgr,
		regionMgr:            regionMgr,
		environmentMgr:       envMgr,
		environmentRegionMgr: envregionmanager.Mgr,
		applicationMgr:       applicationmanager.Mgr,
		groupMgr:             groupmanager.Mgr,
	}

	g, err := groupmanager.Mgr.Create(ctx, &groupmodels.Group{
		Name: "",
		Path: "",
		RegionSelector: `- key: "a"
  values: 
    - "1"`,
	})
	assert.Nil(t, err)
	application, err := applicationmanager.Mgr.Create(ctx, &appmodels.Application{
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
