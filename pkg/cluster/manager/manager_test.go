package manager

import (
	"context"
	"os"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/cluster/models"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	membermanager "github.com/horizoncd/horizon/pkg/member"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	pipelinemodel "github.com/horizoncd/horizon/pkg/pipelinerun/models"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	userdao "github.com/horizoncd/horizon/pkg/user/dao"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
	"github.com/horizoncd/horizon/pkg/util/sets"
	"github.com/stretchr/testify/assert"
)

var (
	db, _     = orm.NewSqliteDB("")
	ctx       context.Context
	mgr       = New(db)
	memberMgr = membermanager.New(db)
	regionMgr = regionmanager.New(db)
)

const secondsInOneDay = 24 * 3600

func TestMain(m *testing.M) {
	currentUser := &usermodels.User{
		Name:  "tony",
		Email: "tony@corp.com",
		Model: global.Model{
			ID: 110,
		},
	}
	// nolint
	db = db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name: currentUser.Name,
		ID:   currentUser.ID,
	}))
	if err := db.AutoMigrate(&models.Cluster{}, &tagmodels.Tag{}, &usermodels.User{},
		&envregionmodels.EnvironmentRegion{}, &regionmodels.Region{}, &membermodels.Member{},
		&pipelinemodel.Pipelinerun{}); err != nil {
		panic(err)
	}
	userDAO := userdao.NewDAO(db)
	_, err := userDAO.Create(ctx, currentUser)
	if err != nil {
		panic(err)
	}
	ctx = context.TODO()
	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: currentUser.Name,
		ID:   currentUser.ID,
	})
	callbacks.RegisterCustomCallbacks(db)
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	userDAO := userdao.NewDAO(db)
	user2, err := userDAO.Create(ctx, &usermodels.User{
		Name:  "leo",
		Email: "leo@corp.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user2)

	currentUser, err := common.UserFromContext(ctx)
	assert.Nil(t, err)

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
		createdBy       = currentUser.GetID()
		updatedBy       = currentUser.GetID()
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
		GitRef:          gitBranch,
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

	count, clustersWithEnvAndRegion, err := mgr.List(ctx,
		q.New(q.KeyWords{
			common.ParamApplicationID:      applicationID,
			common.ClusterQueryEnvironment: []string{environmentName},
		}))
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))
	assert.Equal(t, cluster.Name, clustersWithEnvAndRegion[0].Name)
	assert.Equal(t, environmentName, clustersWithEnvAndRegion[0].EnvironmentName)
	assert.Equal(t, region.Name, clustersWithEnvAndRegion[0].RegionName)

	_, clusters, err := mgr.ListByApplicationID(ctx, applicationID)
	assert.Nil(t, err)
	assert.NotNil(t, clusters)
	assert.Equal(t, 1, len(clusters))

	query := q.New(q.KeyWords{
		common.ClusterQueryName:        "clu",
		common.ClusterQueryEnvironment: environmentName,
	})
	query.PageSize = 1
	query.PageNumber = 1
	count, clustersWithEnvAndRegion, err = mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))

	query = q.New(q.KeyWords{
		common.ClusterQueryName:        "clu",
		common.ClusterQueryEnvironment: environmentName,
		common.ClusterQueryByTemplate:  "javaapp",
		common.ClusterQueryByRelease:   "v1.1.0",
	})
	query.PageSize = 1
	query.PageNumber = 1
	count, clustersWithEnvAndRegion, err = mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, count)
	assert.Equal(t, 1, len(clustersWithEnvAndRegion))

	query = q.New(q.KeyWords{
		common.ClusterQueryName:        "clu",
		common.ClusterQueryEnvironment: environmentName,
		common.ClusterQueryByTemplate:  "node",
	})
	query.PageSize = 1
	query.PageNumber = 1
	count, clustersWithEnvAndRegion, err = mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 0, count)
	assert.Equal(t, 0, len(clustersWithEnvAndRegion))

	query = q.New(q.KeyWords{
		common.ClusterQueryName:        "clu",
		common.ParamApplicationID:      applicationID,
		common.ClusterQueryByUser:      user2.ID,
		common.ClusterQueryEnvironment: environmentName,
	})
	clusterCountForUser, clustersForUser, err := mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, clusterCountForUser)
	for _, cluster := range clustersForUser {
		t.Logf("%v", cluster)
	}

	query = q.New(q.KeyWords{
		common.ClusterQueryName:        "clu",
		common.ParamApplicationID:      applicationID,
		common.ClusterQueryByUser:      user2.ID,
		common.ClusterQueryEnvironment: environmentName,
		common.ClusterQueryByTemplate:  "javaapp",
		common.ClusterQueryByRelease:   "v1.1.0",
	})
	clusterCountForUser, clustersForUser, err = mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, clusterCountForUser)
	for _, cluster := range clustersForUser {
		t.Logf("%v", cluster)
	}

	query = q.New(q.KeyWords{
		common.ClusterQueryName:        "clu",
		common.ParamApplicationID:      applicationID,
		common.ClusterQueryByUser:      user2.ID,
		common.ClusterQueryEnvironment: environmentName,
		common.ClusterQueryByTemplate:  "node",
		common.ClusterQueryByRelease:   "v1.1.0",
	})
	clusterCountForUser, _, err = mgr.List(ctx, query)
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

	clusterGetByID, err = mgr.GetByIDIncludeSoftDelete(ctx, cluster.ID)
	assert.NotNil(t, clusterGetByID)
	assert.Nil(t, err)

	cluster2 := &models.Cluster{
		ApplicationID:   applicationID,
		Name:            "cluster2",
		EnvironmentName: environmentName,
		RegionName:      region.Name,
		Description:     description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitBranch,
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

	query = q.New(q.KeyWords{
		common.ParamApplicationID: applicationID,
		common.ClusterQueryTagSelector: []tagmodels.TagSelector{
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
		},
	})
	query.PageSize = 10
	query.PageNumber = 1
	total, cs, err := mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, len(cs))

	// test ListClusterWithExpiry
	cluster = &models.Cluster{
		ApplicationID:   applicationID,
		Name:            "cluster3",
		EnvironmentName: environmentName,
		RegionName:      region.Name,
		Description:     description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitBranch,
		Template:        template,
		TemplateRelease: templateRelease,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
		Status:          "",
		ExpireSeconds:   1 * secondsInOneDay,
	}
	_, err = mgr.Create(ctx, cluster, []*tagmodels.Tag{
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

	cluster = &models.Cluster{
		ApplicationID:   applicationID,
		Name:            "cluster4",
		EnvironmentName: environmentName,
		RegionName:      region.Name,
		Description:     description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitBranch,
		Template:        template,
		TemplateRelease: templateRelease,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
		Status:          "",
		ExpireSeconds:   5 * secondsInOneDay,
	}
	_, err = mgr.Create(ctx, cluster, []*tagmodels.Tag{
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

	cluster = &models.Cluster{
		ApplicationID:   applicationID,
		Name:            "cluster5",
		EnvironmentName: environmentName,
		RegionName:      region.Name,
		Description:     description,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitBranch,
		Template:        template,
		TemplateRelease: templateRelease,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
		Status:          common.ClusterStatusFreeing,
		ExpireSeconds:   5 * secondsInOneDay,
	}
	_, err = mgr.Create(ctx, cluster, []*tagmodels.Tag{
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

	count, clustersWithEnvAndRegion, err = mgr.List(ctx,
		&q.Query{PageNumber: 1, PageSize: 10})
	assert.Nil(t, err)
	assert.Equal(t, 4, count)
	for _, item := range clustersWithEnvAndRegion {
		t.Logf("%+v", item.Cluster)
	}

	clrs, err := mgr.ListClusterWithExpiry(ctx, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(clrs))
	for _, item := range clrs {
		t.Logf("%+v", item)
	}
	q2 := &q.Query{
		Keywords:   q.KeyWords{common.IDThan: uint(3)},
		Sorts:      nil,
		PageNumber: 1,
		PageSize:   20,
	}
	clrs, err = mgr.ListClusterWithExpiry(ctx, q2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(clrs))
	assert.Equal(t, "cluster4", clrs[0].Name)
}
