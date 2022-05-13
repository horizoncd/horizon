package environmentregion

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/controller/environment"
	"g.hz.netease.com/horizon/core/controller/region"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	"g.hz.netease.com/horizon/pkg/environmentregion/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"github.com/stretchr/testify/assert"
)

var (
	ctx context.Context
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
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
	ctx = context.WithValue(ctx, user.Key(), &userauth.DefaultInfo{
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

	_, err = environment.Ctl.Create(ctx, &environment.CreateEnvironmentRequest{
		Name:          "dev",
		DisplayName:   "DEV",
		DefaultRegion: "hz",
	})
	assert.Nil(t, err)

	er, err := Ctl.CreateEnvironmentRegion(ctx, &CreateEnvironmentRegionRequest{
		EnvironmentName: "dev",
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)

	regions, err := Ctl.ListRegionsByEnvironment(ctx, "dev")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(regions))
	assert.Equal(t, "hz", regions[0].Name)
	assert.Equal(t, "HZ", regions[0].DisplayName)

	regions, err = Ctl.ListRegionsByEnvironment(ctx, "no-exists")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(regions))
}
