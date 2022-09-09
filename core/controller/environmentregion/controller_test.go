package environmentregion

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/environment"
	"g.hz.netease.com/horizon/core/controller/region"
	"g.hz.netease.com/horizon/lib/orm"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	grafanaConfig "g.hz.netease.com/horizon/pkg/config/grafana"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	"g.hz.netease.com/horizon/pkg/environmentregion/models"
	"g.hz.netease.com/horizon/pkg/grafana"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"github.com/stretchr/testify/assert"
	testclient "k8s.io/client-go/kubernetes/fake"
)

var (
	ctx     context.Context
	manager *managerparam.Manager
)

// nolint
func TestMain(m *testing.M) {
	db, _ := orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)
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
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		ID: uint(1),
	})
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	regionCtl := region.NewController(&param.Param{
		Manager: manager,
		GrafanaService: grafana.NewService(grafanaConfig.Config{
			GrafanaURL:                   "",
			DatasourceConfigMapNamespace: "horizon",
		}, testclient.NewSimpleClientset()),
	})
	_, err := regionCtl.Create(ctx, &region.CreateRegionRequest{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)

	environmentCtl := environment.NewController(&param.Param{Manager: manager})
	_, err = environmentCtl.Create(ctx, &environment.CreateEnvironmentRequest{
		Name:        "dev",
		DisplayName: "DEV",
	})
	assert.Nil(t, err)

	ctl := NewController(&param.Param{Manager: manager})
	er, err := ctl.CreateEnvironmentRegion(ctx, &CreateEnvironmentRegionRequest{
		EnvironmentName: "dev",
		RegionName:      "hz",
	})
	assert.Nil(t, err)
	assert.NotNil(t, er)

	ers, err := ctl.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ers))
	assert.Equal(t, "hz", ers[0].RegionName)

	ers, err = ctl.ListByEnvironment(ctx, "dev")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ers))
	assert.Equal(t, "hz", ers[0].RegionName)

	err = ctl.SetEnvironmentRegionToDefault(ctx, er)
	assert.Nil(t, err)

	ers, err = ctl.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ers))
	assert.Equal(t, true, ers[0].IsDefault)

	err = ctl.DeleteByID(ctx, er)
	assert.Nil(t, err)

	ers, err = ctl.ListAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(ers))
}
