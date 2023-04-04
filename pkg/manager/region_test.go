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

package manager

import (
	"context"
	"fmt"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	registrydao "github.com/horizoncd/horizon/pkg/dao"
	applicationmodels "github.com/horizoncd/horizon/pkg/models"
	tagmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func createRegionCtx() (context.Context, RegionManager, EnvironmentRegionManager,
	TagManager, ApplicationRegionManager, ApplicationManager,
	ClusterManager, *gorm.DB) {
	var (
		db, _          = orm.NewSqliteDB("")
		ctx            context.Context
		mgr            = NewRegionManager(db)
		envregionMgr   = NewEnvironmentRegionManager(db)
		tagMgr         = NewTagManager(db)
		appregionMgr   = NewApplicationRegionManager(db)
		applicationMgr = NewApplicationManager(db)
		clusterMgr     = NewClusterManager(db)
	)

	if err := db.AutoMigrate(&applicationmodels.Region{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.Registry{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.Tag{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.Cluster{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.Application{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.Member{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.ApplicationRegion{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:     "1",
		FullName: "1",
		ID:       1,
		Email:    "1",
		Admin:    false,
	})
	return ctx, mgr, envregionMgr, tagMgr, appregionMgr, applicationMgr, clusterMgr, db
}

func TestRegion(t *testing.T) {
	ctx, mgr, _, _, _, _, _, db := createRegionCtx()
	registryDAO := registrydao.NewRegistryDAO(db)
	id, err := registryDAO.Create(ctx, &applicationmodels.Registry{
		Server: "https://harbor1",
		Token:  "asdf",
	})
	assert.Nil(t, err)
	assert.NotNil(t, id)
	rg, err := registryDAO.GetByID(ctx, id)
	assert.Nil(t, err)

	hzRegion, err := mgr.Create(ctx, &applicationmodels.Region{
		Name:          "hz",
		DisplayName:   "HZ",
		Certificate:   "hz-cert",
		IngressDomain: "hz.com",
		PrometheusURL: "hz",
		RegistryID:    id,
	})
	assert.Nil(t, err)
	assert.NotNil(t, hzRegion)

	jdRegion, err := mgr.Create(ctx, &applicationmodels.Region{
		Name:          "jd",
		DisplayName:   "JD",
		Certificate:   "jd-cert",
		IngressDomain: "jd.com",
		PrometheusURL: "jd",
		RegistryID:    id,
	})
	assert.Nil(t, err)
	assert.NotNil(t, jdRegion)

	regions, err := mgr.ListAll(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, regions)
	assert.Equal(t, 2, len(regions))
	assert.Equal(t, "jd", regions[0].Name)
	assert.Equal(t, "hz", regions[1].Name)

	regionEntities, err := mgr.ListRegionEntities(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(regionEntities))

	hzRegionEntity, err := mgr.GetRegionEntity(ctx, "hz")
	assert.Nil(t, err)
	assert.NotNil(t, hzRegionEntity)
	assert.Equal(t, hzRegionEntity.Registry.Server, rg.Server)

	// test updateByID
	err = mgr.UpdateByID(ctx, jdRegion.ID, &applicationmodels.Region{
		Name:          "jd-new",
		DisplayName:   "",
		Server:        "",
		Certificate:   "",
		IngressDomain: "",
		PrometheusURL: "",
		RegistryID:    rg.ID,
		Disabled:      true,
	})
	assert.Nil(t, err)
	regionEntity, err := mgr.GetRegionByID(ctx, jdRegion.ID)
	assert.Nil(t, err)
	assert.Equal(t, "jd", regionEntity.Name)
	assert.Equal(t, true, regionEntity.Disabled)

	region, err := mgr.GetRegionByName(ctx, regionEntity.Name)
	assert.Nil(t, err)
	assert.Equal(t, regionEntity.ID, region.ID)
	assert.Equal(t, true, region.Disabled)
}

func Test_manager_ListByRegionSelectors(t *testing.T) {
	ctx, mgr, _, tagMgr, _, _, _, _ := createRegionCtx()
	r1, err := mgr.Create(ctx, &applicationmodels.Region{
		Name:        "1",
		DisplayName: "1",
		Disabled:    false,
	})
	assert.Nil(t, err)
	r2, err := mgr.Create(ctx, &applicationmodels.Region{
		Name:        "2",
		DisplayName: "2",
		Disabled:    true,
	})
	assert.Nil(t, err)
	r3, err := mgr.Create(ctx, &applicationmodels.Region{
		Name:        "3",
		DisplayName: "3",
		Disabled:    false,
	})
	assert.Nil(t, err)

	err = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r1.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
		{
			Key:   "b",
			Value: "2",
		},
	})
	assert.Nil(t, err)
	err = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r2.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "1",
		},
		{
			Key:   "b",
			Value: "2",
		},
	})
	assert.Nil(t, err)
	err = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r3.ID, []*tagmodels.TagBasic{
		{
			Key:   "a",
			Value: "2",
		},
		{
			Key:   "b",
			Value: "2",
		},
	})
	assert.Nil(t, err)
	type args struct {
		selectors applicationmodels.RegionSelectors
	}
	tests := []struct {
		name    string
		args    args
		want    applicationmodels.RegionParts
		wantErr bool
	}{
		{
			name: "disabled",
			args: args{
				selectors: applicationmodels.RegionSelectors{
					{
						Key:    "b",
						Values: []string{"2"},
					},
				},
			},
			want: applicationmodels.RegionParts{
				{
					Name:        "1",
					DisplayName: "1",
					Disabled:    false,
				},
				{
					Name:        "2",
					DisplayName: "2",
					Disabled:    true,
				},
				{
					Name:        "3",
					DisplayName: "3",
					Disabled:    false,
				},
			},
		},
		{
			name: "oneKeyTwoValues",
			args: args{
				selectors: applicationmodels.RegionSelectors{
					{
						Key:    "a",
						Values: []string{"1", "2"},
					},
				},
			},
			want: applicationmodels.RegionParts{
				{
					Name:        "1",
					DisplayName: "1",
					Disabled:    false,
				},
				{
					Name:        "2",
					DisplayName: "2",
					Disabled:    true,
				},
				{
					Name:        "3",
					DisplayName: "3",
					Disabled:    false,
				},
			},
		},
		{
			name: "twoKeyValuePairs",
			args: args{
				selectors: applicationmodels.RegionSelectors{
					{
						Key:    "a",
						Values: []string{"1"},
					},
					{
						Key:    "b",
						Values: []string{"2"},
					},
				},
			},
			want: applicationmodels.RegionParts{
				{
					Name:        "1",
					DisplayName: "1",
					Disabled:    false,
				},
				{
					Name:        "2",
					DisplayName: "2",
					Disabled:    true,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mgr.ListByRegionSelectors(ctx, tt.args.selectors)
			if (err != nil) != tt.wantErr {
				t.Errorf(fmt.Sprintf("ListByRegionSelectors(%v, %v)", ctx, tt.args.selectors))
			}
			assert.Equalf(t, tt.want, got, "ListByRegionSelectors(%v, %v)", ctx, tt.args.selectors)
		})
	}
}

func Test_manager_DeleteByID(t *testing.T) {
	ctx, mgr, envregionMgr, tagMgr, appregionMgr, applicationMgr, clusterMgr, db := createRegionCtx()
	region, _ := mgr.Create(ctx, &applicationmodels.Region{
		Name:        "1",
		DisplayName: "1",
	})
	application, _ := applicationMgr.Create(ctx, &applicationmodels.Application{
		GroupID: 0,
		Name:    "11",
	}, make(map[string]string))
	_ = appregionMgr.UpsertByApplicationID(ctx, application.ID, []*applicationmodels.ApplicationRegion{
		{
			ID:              0,
			ApplicationID:   application.ID,
			EnvironmentName: "dev",
			RegionName:      region.Name,
		},
	})
	_, _ = envregionMgr.CreateEnvironmentRegion(ctx, &applicationmodels.EnvironmentRegion{
		EnvironmentName: "dev",
		RegionName:      region.Name,
	})

	_, err := clusterMgr.Create(ctx, &applicationmodels.Cluster{
		ApplicationID:   0,
		Name:            "1",
		EnvironmentName: "1",
		RegionName:      region.Name,
	}, []*applicationmodels.Tag{}, map[string]string{})
	assert.Nil(t, err)

	err = mgr.DeleteByID(ctx, region.ID)
	assert.NotNil(t, err)

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&applicationmodels.Cluster{})
	err = mgr.DeleteByID(ctx, region.ID)
	assert.Nil(t, err)

	applicationRegions, _ := appregionMgr.ListByApplicationID(ctx, application.ID)
	assert.Empty(t, applicationRegions)

	regions, _ := envregionMgr.ListAllEnvironmentRegions(ctx)
	assert.Empty(t, regions)

	tags, _ := tagMgr.ListByResourceTypeID(ctx, common.ResourceRegion, region.ID)
	assert.Empty(t, tags)
}
