package cluster

import (
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	applicationmanangermock "github.com/horizoncd/horizon/mock/pkg/application/manager"
	cdmock "github.com/horizoncd/horizon/mock/pkg/cluster/cd"
	clustermanagermock "github.com/horizoncd/horizon/mock/pkg/cluster/manager"
	applicationmodel "github.com/horizoncd/horizon/pkg/application/models"
	"github.com/horizoncd/horizon/pkg/cluster/cd"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	registrymodels "github.com/horizoncd/horizon/pkg/registry/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/stretchr/testify/assert"
)

func testGetClusterStatusV2(t *testing.T) {
	mockCtl := gomock.NewController(t)
	clusterManagerMock := clustermanagermock.NewMockManager(mockCtl)
	appManagerMock := applicationmanangermock.NewMockManager(mockCtl)
	mockCD := cdmock.NewMockCD(mockCtl)
	db, _ := orm.NewSqliteDB("")
	_ = db.AutoMigrate(&regionmodels.Region{}, &registrymodels.Registry{})
	manager := managerparam.InitManager(db)

	regionName := "test"
	status := "expectedStatus"

	c := controller{
		clusterMgr:     clusterManagerMock,
		applicationMgr: appManagerMock,
		regionMgr:      manager.RegionMgr,
		cd:             mockCD,
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

	appManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(3).
		Return(&applicationmodel.Application{}, nil)

	mockCD.EXPECT().GetClusterStateV2(gomock.Any(), gomock.Any()).Times(1).
		Return(&cd.ClusterStateV2{Status: status}, nil)
	mockCD.EXPECT().GetClusterStateV2(gomock.Any(), gomock.Any()).Times(1).
		Return(&cd.ClusterStateV2{Status: status}, nil)
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
