package cluster

import (
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"g.hz.netease.com/horizon/lib/q"
	mockcd "g.hz.netease.com/horizon/mock/pkg/cluster/cd"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	applicationsvc "g.hz.netease.com/horizon/pkg/application/service"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	envmodels "g.hz.netease.com/horizon/pkg/environment/models"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	harbordao "g.hz.netease.com/horizon/pkg/harbor/dao"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	k8sclustermanager "g.hz.netease.com/horizon/pkg/k8scluster/manager"
	k8sclustermodels "g.hz.netease.com/horizon/pkg/k8scluster/models"
	"g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	regionmanager "g.hz.netease.com/horizon/pkg/region/manager"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestListClusterByNameFuzzily(t *testing.T) {
	appMgr := appmanager.Mgr
	groupMgr := groupmanager.Mgr
	clusterMgr := clustermanager.Mgr
	memberMgr := member.Mgr
	envMgr := envmanager.Mgr
	regionMgr := regionmanager.Mgr

	// init data
	var groups []*groupmodels.Group
	for i := 0; i < 5; i++ {
		name := "groupForClusterFuzzily" + strconv.Itoa(i)
		group, err := groupMgr.Create(ctx, &groupmodels.Group{
			Name:     name,
			Path:     name,
			ParentID: 0,
		})
		assert.Nil(t, err)
		assert.NotNil(t, group)
		groups = append(groups, group)
	}

	var applications []*appmodels.Application
	for i := 0; i < 5; i++ {
		group := groups[i]
		name := "appForClusterFuzzily" + strconv.Itoa(i)
		application, err := appMgr.Create(ctx, &appmodels.Application{
			GroupID:         group.ID,
			Name:            name,
			Priority:        "P3",
			GitURL:          "ssh://git.com",
			GitSubfolder:    "/test",
			GitBranch:       "master",
			Template:        "javaapp",
			TemplateRelease: "v1.0.0",
		}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, application)
		applications = append(applications, application)
	}

	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hzFuzzily",
		DisplayName: "HZFuzzily",
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	er, err := envMgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: "testFuzzily",
		RegionName:      "hzFuzzily",
	})
	assert.Nil(t, err)

	for i := 0; i < 5; i++ {
		application := applications[i]
		name := "fuzzilyCluster" + strconv.Itoa(i)
		cluster, err := clusterMgr.Create(ctx, &clustermodels.Cluster{
			ApplicationID:       application.ID,
			Name:                name,
			EnvironmentRegionID: er.ID,
		}, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, cluster)
	}

	c = &controller{
		clusterMgr:     clusterMgr,
		applicationMgr: appMgr,
		applicationSvc: applicationsvc.Svc,
		groupManager:   groupmanager.Mgr,
		memberManager:  memberMgr,
	}

	count, resps, err := c.ListClusterByNameFuzzily(ctx, "", "fuzzilyCluster", nil)
	assert.Nil(t, err)
	assert.Equal(t, 5, count)
	assert.Equal(t, "fuzzilyCluster4", resps[0].Name)
	assert.Equal(t, "fuzzilyCluster3", resps[1].Name)
	assert.Equal(t, "fuzzilyCluster0", resps[4].Name)
	for _, resp := range resps {
		b, _ := json.Marshal(resp)
		t.Logf("%v", string(b))
	}
}

func TestListUserClustersByNameFuzzily(t *testing.T) {
	appMgr := appmanager.Mgr
	groupMgr := groupmanager.Mgr
	clusterMgr := clustermanager.Mgr
	memberMgr := member.Mgr
	envMgr := envmanager.Mgr
	regionMgr := regionmanager.Mgr

	// init data
	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hzUserClustersFuzzily",
		DisplayName: "HZUserClusters",
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	er, err := envMgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: "testUserClustersFuzzily",
		RegionName:      "hzUserClustersFuzzily",
	})
	assert.Nil(t, err)

	var groups []*groupmodels.Group
	for i := 0; i < 5; i++ {
		name := "groupForUserClusterFuzzily" + strconv.Itoa(i)
		group, err := groupMgr.Create(ctx, &groupmodels.Group{
			Name:     name,
			Path:     name,
			ParentID: 0,
		})
		assert.Nil(t, err)
		assert.NotNil(t, group)
		groups = append(groups, group)
	}

	var applications []*appmodels.Application
	for i := 0; i < 5; i++ {
		group := groups[i]
		name := "appForUserClusterFuzzily" + strconv.Itoa(i)
		application, err := appMgr.Create(ctx, &appmodels.Application{
			GroupID:         group.ID,
			Name:            name,
			Priority:        "P3",
			GitURL:          "ssh://git.com",
			GitSubfolder:    "/test",
			GitBranch:       "master",
			Template:        "javaapp",
			TemplateRelease: "v1.0.0",
		}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, application)
		applications = append(applications, application)
	}

	var clusters []*clustermodels.Cluster
	for i := 0; i < 5; i++ {
		application := applications[i]
		name := "userClusterFuzzily" + strconv.Itoa(i)
		cluster, err := clusterMgr.Create(ctx, &clustermodels.Cluster{
			ApplicationID:       application.ID,
			Name:                name,
			EnvironmentRegionID: er.ID,
			GitURL:              "ssh://git@g.hz.netease.com:22222/music-cloud-native/horizon/horizon.git",
		}, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, cluster)
		clusters = append(clusters, cluster)
	}

	_, err = memberMgr.Create(ctx, &membermodels.Member{
		ResourceType: membermodels.TypeGroup,
		ResourceID:   groups[0].ID,
		Role:         "owner",
		MemberType:   membermodels.MemberUser,
		MemberNameID: 1,
	})
	assert.Nil(t, err)

	_, err = memberMgr.Create(ctx, &membermodels.Member{
		ResourceType: membermodels.TypeApplication,
		ResourceID:   applications[1].ID,
		Role:         "owner",
		MemberType:   membermodels.MemberUser,
		MemberNameID: 1,
	})
	assert.Nil(t, err)

	_, err = memberMgr.Create(ctx, &membermodels.Member{
		ResourceType: membermodels.TypeApplicationCluster,
		ResourceID:   clusters[3].ID,
		Role:         "owner",
		MemberType:   membermodels.MemberUser,
		MemberNameID: 1,
	})
	assert.Nil(t, err)

	c = &controller{
		clusterMgr:     clusterMgr,
		applicationMgr: appMgr,
		applicationSvc: applicationsvc.Svc,
		groupManager:   groupmanager.Mgr,
		memberManager:  memberMgr,
	}

	count, resps, err := c.ListUserClusterByNameFuzzily(ctx, er.EnvironmentName, "cluster", nil)
	assert.Nil(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, "userClusterFuzzily3", resps[0].Name)
	assert.Equal(t, "userClusterFuzzily1", resps[1].Name)
	assert.Equal(t, "userClusterFuzzily0", resps[2].Name)
	for _, resp := range resps {
		b, _ := json.Marshal(resp)
		t.Logf("%v", string(b))
	}

	count, resps, err = c.ListUserClusterByNameFuzzily(ctx, er.EnvironmentName, "userCluster", &q.Query{
		PageSize: 2,
	})
	assert.Nil(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, "userClusterFuzzily3", resps[0].Name)
	assert.Equal(t, "userClusterFuzzily1", resps[1].Name)
	for _, resp := range resps {
		b, _ := json.Marshal(resp)
		t.Logf("%v", string(b))
	}
}

func TestController_FreeOrDeleteClusterFailed(t *testing.T) {
	appMgr := appmanager.Mgr
	groupMgr := groupmanager.Mgr
	clusterMgr := clustermanager.Mgr
	envMgr := envmanager.Mgr
	regionMgr := regionmanager.Mgr

	mockCtl := gomock.NewController(t)
	cd := mockcd.NewMockCD(mockCtl)
	cd.EXPECT().DeleteCluster(gomock.Any(), gomock.Any()).Return(errors.New("test")).AnyTimes()

	c = &controller{
		cd:             cd,
		clusterMgr:     clusterMgr,
		applicationMgr: appMgr,
		applicationSvc: applicationsvc.Svc,
		groupManager:   groupmanager.Mgr,
		envMgr:         envmanager.Mgr,
		regionMgr:      regionmanager.Mgr,
	}

	// init data
	k8sCluster, err := k8sclustermanager.Mgr.Create(ctx, &k8sclustermodels.K8SCluster{
		Name: "TestController_FreeOrDeleteClusterFailed",
	})
	assert.Nil(t, err)
	harbor, err := harbordao.NewDAO().Create(ctx, &harbormodels.Harbor{
		Server: "http://127.0.0.1",
	})
	assert.Nil(t, err)
	region, err := regionMgr.Create(ctx, &regionmodels.Region{
		Name:         "TestController_FreeOrDeleteClusterFailed",
		DisplayName:  "TestController_FreeOrDeleteClusterFailed",
		K8SClusterID: k8sCluster.ID,
		HarborID:     harbor.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	er, err := envMgr.CreateEnvironmentRegion(ctx, &envmodels.EnvironmentRegion{
		EnvironmentName: "TestController_FreeOrDeleteClusterFailed",
		RegionName:      region.Name,
	})
	assert.Nil(t, err)

	group, err := groupMgr.Create(ctx, &groupmodels.Group{
		Name:     "TestController_FreeOrDeleteClusterFailed",
		Path:     "/TestController_FreeOrDeleteClusterFailed",
		ParentID: 0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, group)

	application, err := appMgr.Create(ctx, &appmodels.Application{
		GroupID:         group.ID,
		Name:            "TestController_FreeOrDeleteClusterFailed",
		Priority:        "P3",
		GitURL:          "ssh://git.com",
		GitSubfolder:    "/test",
		GitBranch:       "master",
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, application)

	cluster, err := clusterMgr.Create(ctx, &clustermodels.Cluster{
		ApplicationID:       application.ID,
		Name:                "TestController_FreeOrDeleteClusterFailed",
		EnvironmentRegionID: er.ID,
		GitURL:              "",
	}, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)

	// if failed to free, status should be set to empty
	err = c.FreeCluster(ctx, cluster.ID)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	cluster, err = clusterMgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.Equal(t, "", cluster.Status)

	// if failed to delete, status should be set to empty
	err = c.DeleteCluster(ctx, cluster.ID)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	cluster, err = clusterMgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.Equal(t, "", cluster.Status)
}
