package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	"g.hz.netease.com/horizon/pkg/environmentregion/models"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	_, err := regionmanager.Mgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	onlineEnv, err := envmanager.Mgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "online",
		DisplayName: "线上",
	})
	assert.Nil(t, err)
	t.Logf("%v", onlineEnv)

	preEnv, err := envmanager.Mgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "pre",
		DisplayName: "预发",
	})
	assert.Nil(t, err)
	t.Logf("%v", preEnv)

	testEnv, err := envmanager.Mgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "test",
		DisplayName: "测试",
	})
	assert.Nil(t, err)
	t.Logf("%v", testEnv)

	devEnv, err := envmanager.Mgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "dev",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	t.Logf("%v", devEnv)

	devHzEr, err := Mgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, devHzEr)
	t.Logf("%v", devHzEr)

	_, err = Mgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
	})
	assert.NotNil(t, err)
	t.Logf("%v", err)

	regions, err := Mgr.ListRegionsByEnvironment(ctx, devEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, "hz", regions[0].Name)
	t.Logf("%v", regions[0])
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&envmodels.Environment{}); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&models.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
