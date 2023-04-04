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
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	applicationmanangermock "github.com/horizoncd/horizon/mock/pkg/application/manager"
	cdmock "github.com/horizoncd/horizon/mock/pkg/cd"
	clustermanagermock "github.com/horizoncd/horizon/mock/pkg/cluster/manager"
	"github.com/horizoncd/horizon/pkg/cd"
	perror "github.com/horizoncd/horizon/pkg/errors"
	applicationmodel "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/stretchr/testify/assert"
)

func testGetClusterStatusV2(t *testing.T) {
	ctx := context.Background()
	mockCtl := gomock.NewController(t)
	clusterManagerMock := clustermanagermock.NewMockClusterManager(mockCtl)
	appManagerMock := applicationmanangermock.NewMockApplicationManager(mockCtl)
	mockCD := cdmock.NewMockCD(mockCtl)
	db, _ := orm.NewSqliteDB("")
	_ = db.AutoMigrate(&applicationmodel.Region{}, &applicationmodel.Registry{})
	manager := managerparam.InitManager(db)

	regionName := "test"
	status := "expectedStatus"

	c := controller{
		clusterMgr:     clusterManagerMock,
		applicationMgr: appManagerMock,
		regionMgr:      manager.RegionMgr,
		cd:             mockCD,
	}

	_, err := manager.RegistryManager.Create(ctx, &applicationmodel.Registry{
		Model: global.Model{ID: 1},
	})
	assert.Nil(t, err)

	_, err = manager.RegionMgr.Create(ctx, &applicationmodel.Region{
		Model:       global.Model{ID: 1},
		Name:        regionName,
		DisplayName: regionName,
		RegistryID:  1,
	})
	assert.Nil(t, err)

	clusterManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(1).
		Return(&applicationmodel.Cluster{Status: common.ClusterStatusEmpty, RegionName: regionName}, nil)
	clusterManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(1).
		Return(&applicationmodel.Cluster{Status: common.ClusterStatusCreating, RegionName: regionName}, nil)
	clusterManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(1).
		Return(&applicationmodel.Cluster{Status: common.ClusterStatusEmpty, RegionName: regionName}, nil)

	appManagerMock.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(3).
		Return(&applicationmodel.Application{}, nil)

	mockCD.EXPECT().GetClusterState(gomock.Any(), gomock.Any()).Times(1).
		Return(&cd.ClusterStateV2{Status: status}, nil)
	mockCD.EXPECT().GetClusterState(gomock.Any(), gomock.Any()).Times(1).
		Return(&cd.ClusterStateV2{Status: status}, nil)
	mockCD.EXPECT().GetClusterState(gomock.Any(), gomock.Any()).Times(1).
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
