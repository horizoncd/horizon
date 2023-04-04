// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cluster

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	mockcd "github.com/horizoncd/horizon/mock/pkg/cd"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	registrydao "github.com/horizoncd/horizon/pkg/dao"
	appmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/service"
	"github.com/stretchr/testify/assert"
)

func testListClusterByNameFuzzily(t *testing.T) {
	ctx, _, _, _, _,
		_, commitGetter, _, manager, _, _, _, _, _ := createApplicationCtx()
	// init data
	var groups []*appmodels.Group
	for i := 0; i < 5; i++ {
		name := "groupForClusterFuzzily" + strconv.Itoa(i)
		group, err := manager.GroupManager.Create(ctx, &appmodels.Group{
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
		application, err := manager.ApplicationManager.Create(ctx, &appmodels.Application{
			GroupID:         group.ID,
			Name:            name,
			Priority:        "P3",
			GitURL:          "ssh://git.com",
			GitSubfolder:    "/test",
			GitRef:          "master",
			Template:        "javaapp",
			TemplateRelease: "v1.0.0",
		}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, application)
		applications = append(applications, application)
	}

	region, err := manager.RegionMgr.Create(ctx, &appmodels.Region{
		Name:        "hzFuzzily",
		DisplayName: "HZFuzzily",
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	for i := 0; i < 5; i++ {
		application := applications[i]
		name := "fuzzilyCluster" + strconv.Itoa(i)
		cluster, err := manager.ClusterMgr.Create(ctx, &appmodels.Cluster{
			ApplicationID:   application.ID,
			Name:            name,
			EnvironmentName: "testFuzzily",
			RegionName:      "hzFuzzily",
		}, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, cluster)
	}

	c := &controller{
		clusterMgr:     manager.ClusterMgr,
		applicationMgr: manager.ApplicationManager,
		applicationSvc: service.NewApplicationService(service.NewGroupService(manager), manager),
		groupManager:   manager.GroupManager,
		memberManager:  manager.MemberManager,
		eventMgr:       manager.EventManager,
		commitGetter:   commitGetter,
	}

	resps, count, err := c.List(ctx, &q.Query{Keywords: q.KeyWords{common.ClusterQueryName: "fuzzilyCluster"}})
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

func testListUserClustersByNameFuzzily(t *testing.T) {
	ctx, _, _, _, _,
		_, commitGetter, _, manager, _, _, _, _, _ := createApplicationCtx()
	// init data
	region, err := manager.RegionMgr.Create(ctx, &appmodels.Region{
		Name:        "hzUserClustersFuzzily",
		DisplayName: "HZUserClusters",
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	er, err := manager.EnvironmentRegionMgr.CreateEnvironmentRegion(ctx, &appmodels.EnvironmentRegion{
		EnvironmentName: "testUserClustersFuzzily",
		RegionName:      "hzUserClustersFuzzily",
	})
	assert.Nil(t, err)

	var groups []*appmodels.Group
	for i := 0; i < 5; i++ {
		name := "groupForUserClusterFuzzily" + strconv.Itoa(i)
		group, err := manager.GroupManager.Create(ctx, &appmodels.Group{
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
		application, err := manager.ApplicationManager.Create(ctx, &appmodels.Application{
			GroupID:         group.ID,
			Name:            name,
			Priority:        "P3",
			GitURL:          "ssh://git.com",
			GitSubfolder:    "/test",
			GitRef:          "master",
			Template:        "javaapp",
			TemplateRelease: "v1.0.0",
		}, nil)
		assert.Nil(t, err)
		assert.NotNil(t, application)
		applications = append(applications, application)
	}

	var clusters []*appmodels.Cluster
	for i := 0; i < 5; i++ {
		application := applications[i]
		name := "userClusterFuzzily" + strconv.Itoa(i)
		cluster, err := manager.ClusterMgr.Create(ctx, &appmodels.Cluster{
			ApplicationID:   application.ID,
			Name:            name,
			EnvironmentName: "testUserClustersFuzzily",
			RegionName:      "hzUserClustersFuzzily",
			GitURL:          "ssh://git@cloudnative.com:22222/music-cloud-native/horizon/horizon.git",
		}, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, cluster)
		clusters = append(clusters, cluster)
	}

	// nolint
	ctx = context.WithValue(ctx, common.UserContextKey(), &userauth.DefaultInfo{
		Name: "Matt",
		ID:   uint(2),
	})
	_, err = manager.MemberManager.Create(ctx, &appmodels.Member{
		ResourceType: appmodels.TypeGroup,
		ResourceID:   groups[0].ID,
		Role:         "owner",
		MemberType:   appmodels.MemberUser,
		MemberNameID: 2,
	})
	assert.Nil(t, err)

	_, err = manager.MemberManager.Create(ctx, &appmodels.Member{
		ResourceType: appmodels.TypeApplication,
		ResourceID:   applications[1].ID,
		Role:         "owner",
		MemberType:   appmodels.MemberUser,
		MemberNameID: 2,
	})
	assert.Nil(t, err)

	_, err = manager.MemberManager.Create(ctx, &appmodels.Member{
		ResourceType: appmodels.TypeApplicationCluster,
		ResourceID:   clusters[3].ID,
		Role:         "owner",
		MemberType:   appmodels.MemberUser,
		MemberNameID: 2,
	})
	assert.Nil(t, err)

	c := &controller{
		clusterMgr:     manager.ClusterMgr,
		applicationMgr: manager.ApplicationManager,
		applicationSvc: service.NewApplicationService(service.NewGroupService(manager), manager),
		groupManager:   manager.GroupManager,
		memberManager:  manager.MemberManager,
		eventMgr:       manager.EventManager,
		commitGetter:   commitGetter,
	}

	resps, count, err := c.List(ctx,
		&q.Query{
			Keywords: q.KeyWords{
				common.ClusterQueryName:        "cluster",
				common.ClusterQueryEnvironment: er.EnvironmentName,
				common.ClusterQueryByUser:      uint(2),
			}})
	assert.Nil(t, err)
	assert.Equal(t, 3, count)
	assert.Equal(t, "userClusterFuzzily3", resps[0].Name)
	assert.Equal(t, "userClusterFuzzily1", resps[1].Name)
	assert.Equal(t, "userClusterFuzzily0", resps[2].Name)
	for _, resp := range resps {
		b, _ := json.Marshal(resp)
		t.Logf("%v", string(b))
	}

	resps, count, err = c.List(ctx, &q.Query{
		Keywords: q.KeyWords{
			common.ClusterQueryName:        "userCluster",
			common.ClusterQueryEnvironment: er.EnvironmentName,
			common.ClusterQueryByUser:      uint(2),
		},
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

func testListClusterWithExpiry(t *testing.T) {
	ctx, _, _, _, _,
		_, commitGetter, _, manager, _, _, _, _, _ := createApplicationCtx()
	// init data
	clusterInstance := &appmodels.Cluster{
		ApplicationID:   uint(1),
		Name:            "clusterWithExpiry",
		EnvironmentName: "testListClusterWithExpiry",
		RegionName:      "hzListClusterWithExpiry",
		GitURL:          "ssh://git@cloudnative.com:22222/music-cloud-native/horizon/horizon.git",
		Status:          "",
		ExpireSeconds:   secondsInOneDay,
	}

	cluster, err := manager.ClusterMgr.Create(ctx, clusterInstance, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)
	firstClusterID := cluster.ID

	num := 4
	for i := 1; i <= num; i++ {
		clusterInstance.ID = 0
		clusterInstance.Name = "clusterWithExpiry" + strconv.Itoa(i)
		cluster, err := manager.ClusterMgr.Create(ctx, clusterInstance, nil, nil)
		assert.Nil(t, err)
		assert.NotNil(t, cluster)
	}

	c := &controller{
		clusterMgr:     manager.ClusterMgr,
		applicationMgr: manager.ApplicationManager,
		applicationSvc: service.NewApplicationService(service.NewGroupService(manager), manager),
		groupManager:   manager.GroupManager,
		memberManager:  manager.MemberManager,
		eventMgr:       manager.EventManager,
		commitGetter:   commitGetter,
	}

	clusterWithExpiry, err := c.ListClusterWithExpiry(ctx, &q.Query{
		Keywords: q.KeyWords{common.IDThan: int(firstClusterID)},
	})
	assert.Nil(t, err)
	assert.Equal(t, num, len(clusterWithExpiry))
	for _, clr := range clusterWithExpiry {
		t.Logf("%+v", clr)
	}
}

func testControllerFreeOrDeleteClusterFailed(t *testing.T) {
	ctx, _, _, _, _,
		_, _, db, manager, _, _, _, _, _ := createApplicationCtx()
	mockCtl := gomock.NewController(t)
	cd := mockcd.NewMockCD(mockCtl)
	cd.EXPECT().DeleteCluster(gomock.Any(), gomock.Any()).Return(errors.New("test")).AnyTimes()

	c := &controller{
		cd:             cd,
		clusterMgr:     manager.ClusterMgr,
		applicationMgr: manager.ApplicationManager,
		applicationSvc: service.NewApplicationService(service.NewGroupService(manager), manager),
		groupManager:   manager.GroupManager,
		envMgr:         manager.EnvMgr,
		regionMgr:      manager.RegionMgr,
		eventMgr:       manager.EventManager,
	}

	id, err := registrydao.NewRegistryDAO(db).Create(ctx, &appmodels.Registry{
		Server: "http://127.0.0.1",
	})
	assert.Nil(t, err)
	region, err := manager.RegionMgr.Create(ctx, &appmodels.Region{
		Name:        "TestController_FreeOrDeleteClusterFailed",
		DisplayName: "TestController_FreeOrDeleteClusterFailed",
		RegistryID:  id,
	})
	assert.Nil(t, err)
	assert.NotNil(t, region)

	group, err := manager.GroupManager.Create(ctx, &appmodels.Group{
		Name:     "TestController_FreeOrDeleteClusterFailed",
		Path:     "/TestController_FreeOrDeleteClusterFailed",
		ParentID: 0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, group)

	application, err := manager.ApplicationManager.Create(ctx, &appmodels.Application{
		GroupID:         group.ID,
		Name:            "TestController_FreeOrDeleteClusterFailed",
		Priority:        "P3",
		GitURL:          "ssh://git.com",
		GitSubfolder:    "/test",
		GitRef:          "master",
		Template:        "javaapp",
		TemplateRelease: "v1.0.0",
	}, nil)
	assert.Nil(t, err)
	assert.NotNil(t, application)

	cluster, err := manager.ClusterMgr.Create(ctx, &appmodels.Cluster{
		ApplicationID:   application.ID,
		Name:            "TestController_FreeOrDeleteClusterFailed",
		EnvironmentName: "TestController_FreeOrDeleteClusterFailed",
		RegionName:      region.Name,
		GitURL:          "",
	}, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)

	// if failed to free, status should be set to empty
	err = c.FreeCluster(ctx, cluster.ID)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	cluster, err = manager.ClusterMgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.Equal(t, "", cluster.Status)

	// if failed to delete, status should be set to empty
	err = c.DeleteCluster(ctx, cluster.ID, false)
	assert.Nil(t, err)
	time.Sleep(time.Second)
	cluster, err = manager.ClusterMgr.GetByID(ctx, cluster.ID)
	assert.Nil(t, err)
	assert.Equal(t, "", cluster.Status)

	cluster, err = manager.ClusterMgr.Create(ctx, &appmodels.Cluster{
		ApplicationID:   application.ID,
		Name:            "TestController_FreeOrDeleteClusterFailed2",
		EnvironmentName: "TestController_FreeOrDeleteClusterFailed2",
		RegionName:      region.Name,
		GitURL:          "",
	}, nil, nil)
	assert.Nil(t, err)
	assert.NotNil(t, cluster)

	err = c.DeleteCluster(ctx, cluster.ID, true)
	assert.Nil(t, err)
}
