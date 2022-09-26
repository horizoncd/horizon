package environment

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/region"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/environment/models"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"github.com/stretchr/testify/assert"
)

var (
	ctx     context.Context
	manager *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
	if err := db.AutoMigrate(&models.Environment{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		ID: uint(1),
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	regionCtl := region.NewController(&param.Param{
		Manager: manager,
	})
	_, err := regionCtl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	_, err = regionCtl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz-update",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)
	ctl := NewController(&param.Param{
		Manager: manager,
	})
	envs, err := ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(envs))

	devID, err := ctl.Create(ctx, &CreateEnvironmentRequest{
		Name:        "dev",
		DisplayName: "DEV",
		AutoFree:    false,
	})
	assert.Nil(t, err)
	env, err := ctl.GetByID(ctx, devID)
	assert.Nil(t, err)
	assert.Equal(t, "dev", env.Name)
	assert.Equal(t, false, env.AutoFree)

	envByName, err := ctl.GetByName(ctx, env.Name)
	assert.Nil(t, err)
	assert.Equal(t, env, envByName)

	envs, err = ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "dev", envs[0].Name)
	assert.Equal(t, "DEV", envs[0].DisplayName)
	assert.Equal(t, false, envs[0].AutoFree)

	err = ctl.UpdateByID(ctx, devID, &UpdateEnvironmentRequest{
		DisplayName: "DEV-update",
		AutoFree:    true,
	})
	assert.Nil(t, err)

	devID, err = ctl.Create(ctx, &CreateEnvironmentRequest{
		Name:        "online",
		DisplayName: "线上",
		AutoFree:    false,
	})
	assert.Nil(t, err)
	err = ctl.UpdateByID(ctx, devID, &UpdateEnvironmentRequest{
		DisplayName: "线上—online",
		AutoFree:    true,
	})
	assert.NotNil(t, err)
	t.Log(err)

	envs, err = ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(envs))
	assert.Equal(t, "dev", envs[0].Name)
	assert.Equal(t, "DEV-update", envs[0].DisplayName)
	assert.Equal(t, true, envs[0].AutoFree)
	assert.Equal(t, false, envs[1].AutoFree)
}
