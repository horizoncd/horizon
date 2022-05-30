package manager

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	harbordao "g.hz.netease.com/horizon/pkg/harbor/dao"
	harbormodels "g.hz.netease.com/horizon/pkg/harbor/models"
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
	assert.Equal(t, "hz", regions[0].Name)
	assert.Equal(t, "jd", regions[1].Name)

	regionEntities, err := Mgr.ListRegionEntities(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(regionEntities))

	hzRegionEntity, err := Mgr.GetRegionEntity(ctx, "hz")
	assert.Nil(t, err)
	assert.NotNil(t, hzRegionEntity)
	assert.Equal(t, hzRegionEntity.Harbor.Server, harbor.Server)
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
	ctx = orm.NewContext(context.TODO(), db)
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
