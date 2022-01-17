package applicationregion

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/applicationregion/manager"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/config/region"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"

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
		&envmodels.Environment{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	ctx = context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
		Name: "Tony",
		ID:   1,
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	regionMgr := regionmanager.Mgr
	_, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz-test",
		DisplayName: "杭州测试",
	})
	assert.Nil(t, err)

	_, err = regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "杭州",
	})
	assert.Nil(t, err)

	_, err = regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "singapore",
		DisplayName: "新加坡",
	})
	assert.Nil(t, err)

	envMgr := envmanager.Mgr
	for _, env := range []string{"test", "beta", "perf", "pre", "online"} {
		_, err = envMgr.CreateEnvironment(ctx, &envmodels.Environment{
			Name: env,
		})
		assert.Nil(t, err)
	}

	c = &controller{
		regionConfig: &region.Config{
			DefaultRegions: map[string]string{
				"dev":    "hz-test",
				"test":   "hz-test",
				"reg":    "hz-test",
				"perf":   "hz-test",
				"beta":   "hz-test",
				"pre":    "hz",
				"online": "hz",
			},
		},
		mgr:            manager.Mgr,
		regionMgr:      regionMgr,
		environmentMgr: envMgr,
	}

	applicationID := uint(1)

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
