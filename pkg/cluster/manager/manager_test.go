package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Cluster{}, &envmodels.EnvironmentRegion{},
		&regionmodels.Region{}, &membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	var (
		applicationID   = uint(1)
		name            = "cluster"
		description     = "description about cluster"
		gitURL          = "ssh://git@github.com"
		gitSubfolder    = "/"
		gitBranch       = "develop"
		template        = "javaapp"
		templateRelease = "v1.1.0"
		createdBy       = uint(1)
		updatedBy       = uint(1)
	)

	er, err := envmanager.Mgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: "test",
		RegionName:      "hz",
	})
	assert.Nil(t, err)

	region, err := regionmanager.Mgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	cluster := &models.Cluster{
		ApplicationID:       applicationID,
		Name:                name,
		Description:         description,
		GitURL:              gitURL,
		GitSubfolder:        gitSubfolder,
		GitBranch:           gitBranch,
		Template:            template,
		TemplateRelease:     templateRelease,
		EnvironmentRegionID: er.ID,
		CreatedBy:           createdBy,
		UpdatedBy:           updatedBy,
	}

	cluster, err = Mgr.Create(ctx, cluster)
	assert.Nil(t, err)
	t.Logf("%v", cluster)

	cluster.Description = "new Description"
	newCluster, err := Mgr.UpdateByID(ctx, cluster.ID, cluster)
	assert.Nil(t, err)
	assert.Equal(t, cluster.Description, newCluster.Description)

	clusterGetByID, err := Mgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.NotNil(t, clusterGetByID)
	assert.Equal(t, clusterGetByID.Name, cluster.Name)
	t.Logf("%v", clusterGetByID)

	count, clustersWithEnvAndRegion, err := Mgr.ListByApplicationAndEnv(ctx, applicationID,
		er.EnvironmentName, "", nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))
	assert.Equal(t, cluster.Name, clustersWithEnvAndRegion[0].Name)
	assert.Equal(t, er.EnvironmentName, clustersWithEnvAndRegion[0].EnvironmentName)
	assert.Equal(t, er.RegionName, clustersWithEnvAndRegion[0].RegionName)

	clusters, err := Mgr.ListByApplicationID(ctx, applicationID)
	assert.Nil(t, err)
	assert.NotNil(t, clusters)
	assert.Equal(t, 1, len(clusters))

	count, clustersWithEnvAndRegion, err = Mgr.ListByNameFuzzily(ctx, er.EnvironmentName, "clu",
		&q.Query{PageNumber: 1, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))

	exists, err := Mgr.CheckClusterExists(ctx, name)
	assert.Nil(t, err)
	assert.True(t, exists)

	notExists, err := Mgr.CheckClusterExists(ctx, "not-exists")
	assert.Nil(t, err)
	assert.False(t, notExists)

	err = Mgr.DeleteByID(ctx, cluster.ID)
	assert.Nil(t, err)

	clusterGetByID, err = Mgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, clusterGetByID)
	assert.NotNil(t, err)
}
