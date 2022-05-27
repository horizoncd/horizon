package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/cluster/models"
	clustertagmodels "g.hz.netease.com/horizon/pkg/clustertag/models"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	"g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	db = db.Debug()
	// nolint
	db = db.WithContext(context.WithValue(context.Background(), common.ContextUserKey, &userauth.DefaultInfo{
		Name: "tony",
		ID:   110,
	}))
	if err := db.AutoMigrate(&models.Cluster{}, &clustertagmodels.ClusterTag{}, &usermodels.User{},
		&envmodels.EnvironmentRegion{}, &regionmodels.Region{}, &membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	// nolint
	ctx = context.WithValue(ctx, common.ContextUserKey, &userauth.DefaultInfo{
		Name: "tony",
		ID:   110,
	})
	callbacks.RegisterCustomCallbacks(db)
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	userDAO := userdao.NewDAO()
	user1, err := userDAO.Create(ctx, &usermodels.User{
		Name:  "tony",
		Email: "tony@corp.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user1)

	user2, err := userDAO.Create(ctx, &usermodels.User{
		Name:  "leo",
		Email: "leo@corp.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user2)

	var (
		applicationID   = uint(1)
		name            = "cluster"
		description     = "description about cluster"
		gitURL          = "ssh://git@github.com"
		gitSubfolder    = "/"
		gitBranch       = "develop"
		template        = "javaapp"
		templateRelease = "v1.1.0"
		createdBy       = user1.ID
		updatedBy       = user1.ID
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

	cluster, err = Mgr.Create(ctx, cluster, []*clustertagmodels.ClusterTag{
		{
			Key:   "k1",
			Value: "v1",
		},
		{
			Key:   "k2",
			Value: "v2",
		},
	}, map[string]string{user2.Email: role.Owner})
	assert.Nil(t, err)
	t.Logf("%v", cluster)

	clusterMembers, err := member.Mgr.ListDirectMember(ctx, membermodels.TypeApplicationCluster, cluster.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(clusterMembers))
	assert.Equal(t, user2.ID, clusterMembers[1].MemberNameID)
	assert.Equal(t, role.Owner, clusterMembers[1].Role)

	cluster.Description = "new Description"
	newCluster, err := Mgr.UpdateByID(ctx, cluster.ID, cluster)
	assert.Nil(t, err)
	assert.Equal(t, cluster.Description, newCluster.Description)

	clusterGetByID, err := Mgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.NotNil(t, clusterGetByID)
	assert.Equal(t, clusterGetByID.Name, cluster.Name)
	t.Logf("%v", clusterGetByID)

	count, clustersWithEnvAndRegion, err := Mgr.ListByApplicationAndEnvs(ctx, applicationID,
		[]string{er.EnvironmentName}, "", nil)
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

	count, clustersWithEnvAndRegion, err = Mgr.ListByNameFuzzily(ctx, er.EnvironmentName, "clu",
		&q.Query{Keywords: q.KeyWords{"template": "javaapp", "templateRelease": "v1.1.0"}, PageNumber: 1, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))

	count, clustersWithEnvAndRegion, err = Mgr.ListByNameFuzzily(ctx, er.EnvironmentName, "clu",
		&q.Query{Keywords: q.KeyWords{"template": "node"}, PageNumber: 1, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, len(clustersWithEnvAndRegion))

	clusterCountForUser, clustersForUser, err := Mgr.ListUserAuthorizedByNameFuzzily(ctx,
		er.EnvironmentName, "clu", []uint{applicationID}, user2.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, clusterCountForUser)
	for _, cluster := range clustersForUser {
		t.Logf("%v", cluster)
	}

	clusterCountForUser, clustersForUser, err = Mgr.ListUserAuthorizedByNameFuzzily(ctx,
		er.EnvironmentName, "clu", []uint{applicationID}, user2.ID,
		&q.Query{Keywords: q.KeyWords{"template": "javaapp", "templateRelease": "v1.1.0"}})
	assert.Nil(t, err)
	assert.Equal(t, 1, clusterCountForUser)
	for _, cluster := range clustersForUser {
		t.Logf("%v", cluster)
	}

	clusterCountForUser, _, err = Mgr.ListUserAuthorizedByNameFuzzily(ctx,
		er.EnvironmentName, "clu", []uint{applicationID}, user2.ID,
		&q.Query{Keywords: q.KeyWords{"template": "node"}})
	assert.Nil(t, err)
	assert.Equal(t, 0, clusterCountForUser)

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
