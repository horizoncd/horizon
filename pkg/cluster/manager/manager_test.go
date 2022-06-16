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
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	membermanager "g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
	"g.hz.netease.com/horizon/pkg/util/sets"
	"github.com/stretchr/testify/assert"
)

var (
	db, _     = orm.NewSqliteDB("")
	ctx       context.Context
	mgr       = New(db)
	memberMgr = membermanager.New(db)
	regionMgr = regionmanager.New(db)
)

func TestMain(m *testing.M) {
	db = db.Debug()
	// nolint
	db = db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name: "tony",
		ID:   110,
	}))
	if err := db.AutoMigrate(&models.Cluster{}, &tagmodels.Tag{}, &usermodels.User{},
		&envregionmodels.EnvironmentRegion{}, &regionmodels.Region{}, &membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "tony",
		ID:   110,
	})
	callbacks.RegisterCustomCallbacks(db)
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	userDAO := userdao.NewDAO(db)
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
		environmentName = "dev"
		description     = "description about cluster"
		gitURL          = "ssh://git@github.com"
		gitSubfolder    = "/"
		gitBranch       = "develop"
		template        = "javaapp"
		templateRelease = "v1.1.0"
		createdBy       = user1.ID
		updatedBy       = user1.ID
	)

	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	cluster := &models.Cluster{
		ApplicationID:   applicationID,
		EnvironmentName: environmentName,
		RegionName:      region.Name,
		Name:            name,
		Description:     description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitBranch:       gitBranch,
		Template:        template,
		TemplateRelease: templateRelease,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
	}

	cluster, err = mgr.Create(ctx, cluster, []*tagmodels.Tag{
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

	clusterMembers, err := memberMgr.ListDirectMember(ctx, membermodels.TypeApplicationCluster, cluster.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(clusterMembers))
	assert.Equal(t, user2.ID, clusterMembers[1].MemberNameID)
	assert.Equal(t, role.Owner, clusterMembers[1].Role)

	cluster.Description = "new Description"
	newCluster, err := mgr.UpdateByID(ctx, cluster.ID, cluster)
	assert.Nil(t, err)
	assert.Equal(t, cluster.Description, newCluster.Description)

	clusterGetByID, err := mgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.NotNil(t, clusterGetByID)
	assert.Equal(t, clusterGetByID.Name, cluster.Name)
	t.Logf("%v", clusterGetByID)

	count, clustersWithEnvAndRegion, err := mgr.ListByApplicationEnvsTags(ctx, applicationID,
		[]string{environmentName}, "", nil, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))
	assert.Equal(t, cluster.Name, clustersWithEnvAndRegion[0].Name)
	assert.Equal(t, environmentName, clustersWithEnvAndRegion[0].EnvironmentName)
	assert.Equal(t, region.Name, clustersWithEnvAndRegion[0].RegionName)

	clusters, err := mgr.ListByApplicationID(ctx, applicationID)
	assert.Nil(t, err)
	assert.NotNil(t, clusters)
	assert.Equal(t, 1, len(clusters))

	count, clustersWithEnvAndRegion, err = mgr.ListByNameFuzzily(ctx, environmentName, "clu",
		&q.Query{PageNumber: 1, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))

	count, clustersWithEnvAndRegion, err = mgr.ListByNameFuzzily(ctx, environmentName, "clu",
		&q.Query{Keywords: q.KeyWords{"template": "javaapp", "templateRelease": "v1.1.0"}, PageNumber: 1, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))

	count, clustersWithEnvAndRegion, err = mgr.ListByNameFuzzily(ctx, environmentName, "clu",
		&q.Query{Keywords: q.KeyWords{"template": "node"}, PageNumber: 1, PageSize: 1})
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, len(clustersWithEnvAndRegion))

	clusterCountForUser, clustersForUser, err := mgr.ListUserAuthorizedByNameFuzzily(ctx,
		environmentName, "clu", []uint{applicationID}, user2.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, clusterCountForUser)
	for _, cluster := range clustersForUser {
		t.Logf("%v", cluster)
	}

	clusterCountForUser, clustersForUser, err = mgr.ListUserAuthorizedByNameFuzzily(ctx,
		environmentName, "clu", []uint{applicationID}, user2.ID,
		&q.Query{Keywords: q.KeyWords{"template": "javaapp", "templateRelease": "v1.1.0"}})
	assert.Nil(t, err)
	assert.Equal(t, 1, clusterCountForUser)
	for _, cluster := range clustersForUser {
		t.Logf("%v", cluster)
	}

	clusterCountForUser, _, err = mgr.ListUserAuthorizedByNameFuzzily(ctx,
		environmentName, "clu", []uint{applicationID}, user2.ID,
		&q.Query{Keywords: q.KeyWords{"template": "node"}})
	assert.Nil(t, err)
	assert.Equal(t, 0, clusterCountForUser)

	exists, err := mgr.CheckClusterExists(ctx, name)
	assert.Nil(t, err)
	assert.True(t, exists)

	notExists, err := mgr.CheckClusterExists(ctx, "not-exists")
	assert.Nil(t, err)
	assert.False(t, notExists)

	err = mgr.DeleteByID(ctx, cluster.ID)
	assert.Nil(t, err)

	clusterGetByID, err = mgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, clusterGetByID)
	assert.NotNil(t, err)

	cluster2 := &models.Cluster{
		ApplicationID:   applicationID,
		Name:            "cluster2",
		EnvironmentName: environmentName,
		RegionName:      region.Name,
		Description:     description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitBranch:       gitBranch,
		Template:        template,
		TemplateRelease: templateRelease,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
	}
	_, err = mgr.Create(ctx, cluster2, []*tagmodels.Tag{
		{
			Key:   "k1",
			Value: "v3",
		},
		{
			Key:   "k3",
			Value: "v3",
		},
	}, map[string]string{user2.Email: role.Owner})
	assert.Nil(t, err)
	t.Logf("%v", cluster)

	total, cs, err := mgr.ListByApplicationEnvsTags(ctx, applicationID, nil, "", &q.Query{
		PageNumber: 1,
		PageSize:   10,
	}, []tagmodels.TagSelector{
		{
			Key:      "k1",
			Operator: tagmodels.In,
			Values:   sets.NewString("v1", "v3"),
		},
		{
			Key:      "k3",
			Operator: tagmodels.In,
			Values:   sets.NewString("v3"),
		},
	})
	assert.Nil(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, len(cs))
}
