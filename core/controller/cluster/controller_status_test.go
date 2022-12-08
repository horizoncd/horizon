package cluster

import (
	"testing"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/orm"
	cdmock "g.hz.netease.com/horizon/mock/pkg/cluster/cd"
	clustermanagermock "g.hz.netease.com/horizon/mock/pkg/cluster/manager"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	registrymodels "g.hz.netease.com/horizon/pkg/registry/models"
	"g.hz.netease.com/horizon/pkg/server/global"
	"github.com/argoproj/gitops-engine/pkg/health"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func testGetClusterStatusV2(t *testing.T) {
	mockCtl := gomock.NewController(t)
	clusterManagerMock := clustermanagermock.NewMockManager(mockCtl)
	mockCD := cdmock.NewMockCD(mockCtl)
	db, _ := orm.NewSqliteDB("")
	_ = db.AutoMigrate(&regionmodels.Region{}, &registrymodels.Registry{})
	manager := managerparam.InitManager(db)

	regionName := "test"
	status := "expectedStatus"

	c := controller{
		clusterMgr: clusterManagerMock,
		regionMgr:  manager.RegionMgr,
		cd:         mockCD,
	}

	_, err := manager.RegistryManager.Create(ctx, &registrymodels.Registry{
		Model: global.Model{ID: 1},
	})
	assert.Nil(t, err)

	_, err = manager.RegionMgr.Create(ctx, &regionmodels.Region{
		Model:       global.Model{ID: 1},
		Name:        regionName,
		DisplayName: regionName,
		RegistryID:  1,
	})
	assert.Nil(t, err)

	clusterManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(1).
		Return(&clustermodels.Cluster{Status: common.ClusterStatusEmpty, RegionName: regionName}, nil)
	clusterManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(1).
		Return(&clustermodels.Cluster{Status: common.ClusterStatusCreating, RegionName: regionName}, nil)
	clusterManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(1).
		Return(&clustermodels.Cluster{Status: common.ClusterStatusEmpty, RegionName: regionName}, nil)

	mockCD.EXPECT().GetClusterStateV2(gomock.Any(), gomock.Any()).Times(1).
		Return(&cd.ClusterStateV2{Status: health.HealthStatusCode(status)}, nil)
	mockCD.EXPECT().GetClusterStateV2(gomock.Any(), gomock.Any()).Times(1).
		Return(&cd.ClusterStateV2{Status: health.HealthStatusCode(status)}, nil)
	mockCD.EXPECT().GetClusterStateV2(gomock.Any(), gomock.Any()).Times(1).
		Return(nil, perror.Wrap(herrors.NewErrNotFound(herrors.ApplicationInArgo, ""), ""))

	resp, err := c.GetClusterStatusV2(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, status, resp.Status)

	resp, err = c.GetClusterStatusV2(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, common.ClusterStatusCreating, resp.Status)

	resp, err = c.GetClusterStatusV2(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, _notFound, resp.Status)
}
