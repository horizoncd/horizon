package manager

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/middleware/user"
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
	harbordao "g.hz.netease.com/horizon/pkg/harbor/dao"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/region/models"
	tagmanager "g.hz.netease.com/horizon/pkg/tag/manager"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	harborDAO := harbordao.NewDAO()
	harbor, err := harborDAO.Create(ctx, &harbormodels.Harbor{
		Server:          "https://harbor1",
		Token:           "asdf",
		PreheatPolicyID: 1,
	})
	assert.Nil(t, err)
	assert.NotNil(t, harbor)

	hzRegion, err := Mgr.Create(ctx, &models.Region{
		Name:          "hz",
		DisplayName:   "HZ",
		Certificate:   "hz-cert",
		IngressDomain: "hz.com",
		HarborID:      harbor.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, hzRegion)

	jdRegion, err := Mgr.Create(ctx, &models.Region{
		Name:          "jd",
		DisplayName:   "JD",
		Certificate:   "jd-cert",
		IngressDomain: "jd.com",
		HarborID:      harbor.ID,
	})
	assert.Nil(t, err)
	assert.NotNil(t, jdRegion)

	regions, err := Mgr.ListAll(ctx)
	assert.Nil(t, err)
	assert.NotNil(t, regions)
	assert.Equal(t, 2, len(regions))
	assert.Equal(t, "jd", regions[0].Name)
	assert.Equal(t, "hz", regions[1].Name)

	regionEntities, err := Mgr.ListRegionEntities(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(regionEntities))

	hzRegionEntity, err := Mgr.GetRegionEntity(ctx, "hz")
	assert.Nil(t, err)
	assert.NotNil(t, hzRegionEntity)
	assert.Equal(t, hzRegionEntity.Harbor.Server, harbor.Server)

	// test updateByID
	err = Mgr.UpdateByID(ctx, jdRegion.ID, &models.Region{
		Name:          "jd-new",
		DisplayName:   "",
		Server:        "",
		Certificate:   "",
		IngressDomain: "",
		HarborID:      harbor.ID,
		Disabled:      true,
	})
	assert.Nil(t, err)
	regions, _ = Mgr.ListAll(ctx)
	assert.Equal(t, "jd", regions[0].Name)
	assert.Equal(t, true, regions[0].Disabled)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Region{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&harbormodels.Harbor{}); err != nil {
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
	ctx = orm.NewContext(context.TODO(), db)
	ctx = user.WithContext(ctx, &userauth.DefaultInfo{
		Name:     "1",
		FullName: "1",
		ID:       1,
		Email:    "1",
		Admin:    false,
	})
	os.Exit(m.Run())
}

func Test_manager_ListByRegionSelectors(t *testing.T) {
	r1, err := Mgr.Create(ctx, &models.Region{
		Name:        "1",
		DisplayName: "1",
		Disabled:    false,
	})
	assert.Nil(t, err)
	r2, err := Mgr.Create(ctx, &models.Region{
		Name:        "2",
		DisplayName: "2",
		Disabled:    true,
	})
	assert.Nil(t, err)
	r3, err := Mgr.Create(ctx, &models.Region{
		Name:        "3",
		DisplayName: "3",
		Disabled:    false,
	})
	assert.Nil(t, err)

	err = tagmanager.Mgr.UpsertByResourceTypeID(ctx, tagmodels.TypeRegion, r1.ID, []*tagmodels.Tag{
		{
			ResourceID:   r1.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "a",
			Value:        "1",
		},
		{
			ResourceID:   r1.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "b",
			Value:        "2",
		},
	})
	assert.Nil(t, err)
	err = tagmanager.Mgr.UpsertByResourceTypeID(ctx, tagmodels.TypeRegion, r2.ID, []*tagmodels.Tag{
		{
			ResourceID:   r2.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "a",
			Value:        "1",
		},
		{
			ResourceID:   r2.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "b",
			Value:        "2",
		},
	})
	assert.Nil(t, err)
	err = tagmanager.Mgr.UpsertByResourceTypeID(ctx, tagmodels.TypeRegion, r3.ID, []*tagmodels.Tag{
		{
			ResourceID:   r3.ID,
			ResourceType: tagmodels.TypeRegion,
			Key:          "a",
			Value:        "2",
		},
		{
			ResourceID:   r3.ID,
			ResourceType: tagmodels.TypeRegion,
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
				},
				{
					Name:        "3",
					DisplayName: "3",
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
				},
				{
					Name:        "3",
					DisplayName: "3",
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
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mgr.ListByRegionSelectors(ctx, tt.args.selectors)
			if (err != nil) != tt.wantErr {
				t.Errorf(fmt.Sprintf("ListByRegionSelectors(%v, %v)", ctx, tt.args.selectors))
			}
			assert.Equalf(t, tt.want, got, "ListByRegionSelectors(%v, %v)", ctx, tt.args.selectors)
		})
	}
}

func Test_manager_DeleteByID(t *testing.T) {
	region, _ := Mgr.Create(ctx, &models.Region{
		Name:        "1",
		DisplayName: "1",
	})
	application, _ := applicationmanager.Mgr.Create(ctx, &applicationmodels.Application{
		GroupID: 0,
		Name:    "11",
	}, make(map[string]string))
	_ = appregionmanager.Mgr.UpsertByApplicationID(ctx, application.ID, []*appregionmodels.ApplicationRegion{
		{
			ID:              0,
			ApplicationID:   application.ID,
			EnvironmentName: "dev",
			RegionName:      region.Name,
		},
	})
	_, _ = envregionmanager.Mgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: "dev",
		RegionName:      region.Name,
	})

	_, err := clustermanager.Mgr.Create(ctx, &clustermodels.Cluster{
		ApplicationID:   0,
		Name:            "1",
		EnvironmentName: "1",
		RegionName:      region.Name,
	}, []*tagmodels.Tag{}, map[string]string{})
	assert.Nil(t, err)

	err = Mgr.DeleteByID(ctx, region.ID)
	assert.NotNil(t, err)

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&clustermodels.Cluster{})
	err = Mgr.DeleteByID(ctx, region.ID)
	assert.Nil(t, err)

	applicationRegions, _ := appregionmanager.Mgr.ListByApplicationID(ctx, application.ID)
	assert.Empty(t, applicationRegions)

	regions, _ := envregionmanager.Mgr.ListAllEnvironmentRegions(ctx)
	assert.Empty(t, regions)

	tags, _ := tagmanager.Mgr.ListByResourceTypeID(ctx, tagmodels.TypeRegion, region.ID)
	assert.Empty(t, tags)
}
