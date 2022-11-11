package manager

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
	appregionmanager "g.hz.netease.com/horizon/pkg/applicationregion/manager"
	appregionmodels "g.hz.netease.com/horizon/pkg/applicationregion/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	envregionmanager "g.hz.netease.com/horizon/pkg/environmentregion/manager"
	envregionmodels "g.hz.netease.com/horizon/pkg/environmentregion/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/region/models"
	registrydao "g.hz.netease.com/horizon/pkg/registry/dao"
	registrymodels "g.hz.netease.com/horizon/pkg/registry/models"
	tagmanager "g.hz.netease.com/horizon/pkg/tag/manager"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db, _          = orm.NewSqliteDB("")
	ctx            context.Context
	mgr            = New(db)
	envregionMgr   = envregionmanager.New(db)
	tagMgr         = tagmanager.New(db)
	appregionMgr   = appregionmanager.New(db)
	applicationMgr = applicationmanager.New(db)
	clusterMgr     = clustermanager.New(db)
)

func Test(t *testing.T) {
	registryDAO := registrydao.NewDAO(db)
	id, err := registryDAO.Create(ctx, &registrymodels.Registry{
		Server: "https://harbor1",
		Token:  "asdf",
	})
	assert.Nil(t, err)
	assert.NotNil(t, id)
	rg, err := registryDAO.GetByID(ctx, id)
	assert.Nil(t, err)

	hzRegion, err := mgr.Create(ctx, &models.Region{
		Name:          "hz",
		DisplayName:   "HZ",
		Certificate:   "hz-cert",
		IngressDomain: "hz.com",
		PrometheusURL: "hz",
		RegistryID:    id,
	})
	assert.Nil(t, err)
	assert.NotNil(t, hzRegion)

	jdRegion, err := mgr.Create(ctx, &models.Region{
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
	err = mgr.UpdateByID(ctx, jdRegion.ID, &models.Region{
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

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.Region{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&registrymodels.Registry{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&tagmodels.Tag{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&envregionmodels.EnvironmentRegion{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&clustermodels.Cluster{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&applicationmodels.Application{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&membermodels.Member{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&appregionmodels.ApplicationRegion{}); err != nil {
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
	os.Exit(m.Run())
}

func Test_manager_ListByRegionSelectors(t *testing.T) {
	r1, err := mgr.Create(ctx, &models.Region{
		Name:        "1",
		DisplayName: "1",
		Disabled:    false,
	})
	assert.Nil(t, err)
	r2, err := mgr.Create(ctx, &models.Region{
		Name:        "2",
		DisplayName: "2",
		Disabled:    true,
	})
	assert.Nil(t, err)
	r3, err := mgr.Create(ctx, &models.Region{
		Name:        "3",
		DisplayName: "3",
		Disabled:    false,
	})
	assert.Nil(t, err)

	err = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r1.ID, []*tagmodels.Tag{
		{
			ResourceID:   r1.ID,
			ResourceType: common.ResourceRegion,
			Key:          "a",
			Value:        "1",
		},
		{
			ResourceID:   r1.ID,
			ResourceType: common.ResourceRegion,
			Key:          "b",
			Value:        "2",
		},
	})
	assert.Nil(t, err)
	err = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r2.ID, []*tagmodels.Tag{
		{
			ResourceID:   r2.ID,
			ResourceType: common.ResourceRegion,
			Key:          "a",
			Value:        "1",
		},
		{
			ResourceID:   r2.ID,
			ResourceType: common.ResourceRegion,
			Key:          "b",
			Value:        "2",
		},
	})
	assert.Nil(t, err)
	err = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r3.ID, []*tagmodels.Tag{
		{
			ResourceID:   r3.ID,
			ResourceType: common.ResourceRegion,
			Key:          "a",
			Value:        "2",
		},
		{
			ResourceID:   r3.ID,
			ResourceType: common.ResourceRegion,
			Key:          "b",
			Value:        "2",
		},
	})
	assert.Nil(t, err)
	type args struct {
		selectors groupmodels.RegionSelectors
	}
	tests := []struct {
		name    string
		args    args
		want    models.RegionParts
		wantErr bool
	}{
		{
			name: "disabled",
			args: args{
				selectors: groupmodels.RegionSelectors{
					{
						Key:    "b",
						Values: []string{"2"},
					},
				},
			},
			want: models.RegionParts{
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
				selectors: groupmodels.RegionSelectors{
					{
						Key:    "a",
						Values: []string{"1", "2"},
					},
				},
			},
			want: models.RegionParts{
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
				selectors: groupmodels.RegionSelectors{
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
			want: models.RegionParts{
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
	region, _ := mgr.Create(ctx, &models.Region{
		Name:        "1",
		DisplayName: "1",
	})
	application, _ := applicationMgr.Create(ctx, &applicationmodels.Application{
		GroupID: 0,
		Name:    "11",
	}, make(map[string]string))
	_ = appregionMgr.UpsertByApplicationID(ctx, application.ID, []*appregionmodels.ApplicationRegion{
		{
			ID:              0,
			ApplicationID:   application.ID,
			EnvironmentName: "dev",
			RegionName:      region.Name,
		},
	})
	_, _ = envregionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "dev",
		RegionName:      region.Name,
	})

	_, err := clusterMgr.Create(ctx, &clustermodels.Cluster{
		ApplicationID:   0,
		Name:            "1",
		EnvironmentName: "1",
		RegionName:      region.Name,
	}, []*tagmodels.Tag{}, map[string]string{})
	assert.Nil(t, err)

	err = mgr.DeleteByID(ctx, region.ID)
	assert.NotNil(t, err)

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&clustermodels.Cluster{})
	err = mgr.DeleteByID(ctx, region.ID)
	assert.Nil(t, err)

	applicationRegions, _ := appregionMgr.ListByApplicationID(ctx, application.ID)
	assert.Empty(t, applicationRegions)

	regions, _ := envregionMgr.ListAllEnvironmentRegions(ctx)
	assert.Empty(t, regions)

	tags, _ := tagMgr.ListByResourceTypeID(ctx, common.ResourceRegion, region.ID)
	assert.Empty(t, tags)
}
