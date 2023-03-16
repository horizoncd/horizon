package manager

import (
	"context"
	"os"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	envmanager "github.com/horizoncd/horizon/pkg/environment/manager"
	envmodels "github.com/horizoncd/horizon/pkg/environment/models"
	"github.com/horizoncd/horizon/pkg/environmentregion/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"

	"github.com/stretchr/testify/assert"
)

var (
	db, _     = orm.NewSqliteDB("")
	ctx       context.Context
	regionMgr = regionmanager.New(db)
	envMgr    = envmanager.New(db)
	mgr       = New(db)
)

func Test(t *testing.T) {
	_, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	onlineEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "online",
		DisplayName: "线上",
	})
	assert.Nil(t, err)
	t.Logf("%v", onlineEnv)

	preEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "pre",
		DisplayName: "预发",
	})
	assert.Nil(t, err)
	t.Logf("%v", preEnv)

	testEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "test",
		DisplayName: "测试",
	})
	assert.Nil(t, err)
	t.Logf("%v", testEnv)

	devEnv, err := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "dev",
		DisplayName: "开发",
	})
	assert.Nil(t, err)
	t.Logf("%v", devEnv)

	devHzEr, err := mgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
		IsDefault:       true,
	})
	assert.Nil(t, err)
	assert.NotNil(t, devHzEr)
	t.Logf("%v", devHzEr)

	_, err = mgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
	})
	assert.NotNil(t, err)
	t.Logf("%v", err)

	regions, err := mgr.ListByEnvironment(ctx, devEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, "hz", regions[0].RegionName)
	t.Logf("%v", regions[0])

	// test SetEnvironmentRegionToDefaultByID
	r2, _ := mgr.CreateEnvironmentRegion(ctx, &models.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz1",
	})

	err = mgr.SetEnvironmentRegionToDefaultByID(ctx, r2.ID)
	assert.Nil(t, err)
	r2New, err := mgr.GetEnvironmentRegionByID(ctx, r2.ID)
	assert.Nil(t, err)
	assert.Equal(t, r2New.IsDefault, true)
	r2New, err = mgr.GetDefaultRegionByEnvironment(ctx, devEnv.Name)
	assert.Nil(t, err)
	assert.Equal(t, r2New.IsDefault, true)

	devHzErNew, err := mgr.GetEnvironmentRegionByID(ctx, devHzEr.ID)
	assert.Nil(t, err)
	assert.Equal(t, devHzErNew.IsDefault, false)
	devHzErNew, err = mgr.GetByEnvironmentAndRegion(ctx, devEnv.Name, "hz")
	assert.Nil(t, err)
	assert.Equal(t, devHzErNew.IsDefault, false)

	// test deleteByID
	err = mgr.DeleteByID(ctx, devHzErNew.ID)
	assert.Nil(t, err)
	_, err = mgr.GetEnvironmentRegionByID(ctx, devHzEr.ID)
	_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&envmodels.Environment{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&models.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}
