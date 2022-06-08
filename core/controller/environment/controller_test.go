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
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"github.com/stretchr/testify/assert"
)

var (
	ctx context.Context
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Environment{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		ID: uint(1),
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	_, err := region.Ctl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	_, err = region.Ctl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz-update",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)
	envs, err := Ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(envs))

	devID, err := Ctl.Create(ctx, &CreateEnvironmentRequest{
		Name:        "dev",
		DisplayName: "DEV",
	})
	assert.Nil(t, err)

	envs, err = Ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "dev", envs[0].Name)
	assert.Equal(t, "DEV", envs[0].DisplayName)

	err = Ctl.UpdateByID(ctx, devID, &UpdateEnvironmentRequest{
		DisplayName: "DEV-update",
	})
	assert.Nil(t, err)

	envs, err = Ctl.ListEnvironments(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(envs))
	assert.Equal(t, "dev", envs[0].Name)
	assert.Equal(t, "DEV-update", envs[0].DisplayName)
}
