package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Cluster{}); err != nil {
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

	cluster := &models.Cluster{
		ApplicationID:       applicationID,
		Name:                name,
		Description:         description,
		GitURL:              gitURL,
		GitSubfolder:        gitSubfolder,
		GitBranch:           gitBranch,
		Template:            template,
		TemplateRelease:     templateRelease,
		EnvironmentRegionID: uint(2),
		CreatedBy:           createdBy,
		UpdatedBy:           updatedBy,
	}

	cluster, err := Mgr.Create(ctx, cluster)
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

	clusters, err := Mgr.ListByApplication(ctx, applicationID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(clusters))
	assert.Equal(t, cluster.Name, clusters[0].Name)

	exists, err := Mgr.CheckClusterExists(ctx, name)
	assert.Nil(t, err)
	assert.True(t, exists)

	notExists, err := Mgr.CheckClusterExists(ctx, "not-exists")
	assert.Nil(t, err)
	assert.False(t, notExists)
}
