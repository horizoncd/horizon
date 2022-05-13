package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/environment/models"
	models2 "g.hz.netease.com/horizon/pkg/environmentregion/models"
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
		Name:          "online",
		DisplayName:   "线上",
		DefaultRegion: "hz",
	})
	assert.Nil(t, err)
	t.Logf("%v", onlineEnv)
	err = Mgr.UpdateByID(ctx, onlineEnv.ID, &models.Environment{
		Name:          "online-update",
		DisplayName:   "线上-update",
		DefaultRegion: "hz-update",
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
	assert.Equal(t, envs[3].DefaultRegion, "hz-update")
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Environment{}); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&models2.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
