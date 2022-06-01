package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	appregionmanager "g.hz.netease.com/horizon/pkg/applicationregion/manager"
	appregionmodels "g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/environment/models"
	envregionmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	envregion "g.hz.netease.com/horizon/pkg/environmentregion/models"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
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

	_, err = regionmanager.Mgr.Create(ctx, &regionmodels.Region{
		Name:        "hz-update",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	onlineEnv, err := Mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "online",
		DisplayName: "线上",
	})
	assert.Nil(t, err)
	t.Logf("%v", onlineEnv)
	err = Mgr.UpdateByID(ctx, onlineEnv.ID, &models.Environment{
		Name:        "online-update",
		DisplayName: "线上-update",
	})
	assert.Nil(t, err)

	preEnv, err := Mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "pre",
		DisplayName: "预发",
	})
	assert.Nil(t, err)
	t.Logf("%v", preEnv)

	testEnv, err := Mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "test",
		DisplayName: "测试",
	})
	assert.Nil(t, err)
	t.Logf("%v", testEnv)

	devEnv, err := Mgr.CreateEnvironment(ctx, &models.Environment{
		Name:        "dev",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	t.Logf("%v", devEnv)

	envs, err := Mgr.ListAllEnvironment(ctx)
	assert.Nil(t, err)
	assert.Equal(t, len(envs), 4)
	t.Logf("%v", envs[0])
	t.Logf("%v", envs[1])
	t.Logf("%v", envs[2])
	t.Logf("%v", envs[3])
	assert.Equal(t, envs[3].Name, "online")
	assert.Equal(t, envs[3].DisplayName, "线上-update")

	err = appregionmanager.Mgr.UpsertByApplicationID(ctx, uint(1), []*appregionmodels.ApplicationRegion{
		{
			ID:              0,
			ApplicationID:   uint(1),
			EnvironmentName: "dev",
			RegionName:      "",
		},
	})
	assert.Nil(t, err)
	_, err = envregionmanager.Mgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "dev",
		RegionName:      "",
	})
	assert.Nil(t, err)

	err = Mgr.DeleteByID(ctx, devEnv.ID)
	assert.Nil(t, err)

	applicationRegions, _ := appregionmanager.Mgr.ListByApplicationID(ctx, uint(1))
	assert.Empty(t, applicationRegions)

	regions, _ := envregionmanager.Mgr.ListRegionsByEnvironment(ctx, devEnv.Name)
	assert.Empty(t, regions)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
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
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
