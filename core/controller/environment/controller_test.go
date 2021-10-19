package environment

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/environment/manager"
	"g.hz.netease.com/horizon/pkg/environment/models"
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
	if err := db.AutoMigrate(&models.Environment{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&models.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	ctx = context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
		ID: uint(1),
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	envMgr := manager.Mgr
	regionMgr := regionmanager.Mgr
	c = &controller{
		envMgr: envMgr,
	}
	envs, err := c.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(envs))

	devEnv, err := envMgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "dev",
		DisplayName: "DEV",
	})
	assert.Nil(t, err)

	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:         "hz",
		DisplayName:  "HZ",
		K8SClusterID: 0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	er, err := envMgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)

	envs, err = c.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "dev", envs[0].Name)
	assert.Equal(t, "DEV", envs[0].DisplayName)

	regions, err := c.ListRegionsByEnvironment(ctx, devEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, "hz", regions[0].Name)
	assert.Equal(t, "HZ", regions[0].DisplayName)

	regions, err = c.ListRegionsByEnvironment(ctx, "no-exists")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(regions))
}
