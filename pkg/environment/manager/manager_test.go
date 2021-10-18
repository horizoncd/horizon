package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/environment/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	envs, err := Mgr.ListAllEnvironment(ctx)
	assert.Nil(t, err)
	assert.Equal(t, len(envs), 0)

	testEnv, err := Mgr.CreateEnvironment(ctx, &models.Environment{
		Env:  "test",
		Name: "测试",
	})
	assert.Nil(t, err)
	t.Logf("%v", testEnv)

	devEnv, err := Mgr.CreateEnvironment(ctx, &models.Environment{
		Env:  "dev",
		Name: "开发",
	})
	assert.Nil(t, err)
	t.Logf("%v", devEnv)

	envs, err = Mgr.ListAllEnvironment(ctx)
	assert.Nil(t, err)
	assert.Equal(t, len(envs), 2)
	t.Logf("%v", envs[0])
	t.Logf("%v", envs[1])

	devHzEr, err := Mgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		Env:    devEnv.Env,
		Region: "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, devHzEr)
	t.Logf("%v", devHzEr)

	_, err = Mgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		Env:    devEnv.Env,
		Region: "hz",
	})
	assert.NotNil(t, err)
	t.Logf("%v", err)

	regions, err := Mgr.ListRegionsByEnvironment(ctx, devEnv.Env)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, "hz", regions[0])

}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Environment{}); err != nil {
		panic(err)
	}

	if err := db.AutoMigrate(&models.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
