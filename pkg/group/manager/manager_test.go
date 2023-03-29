package manager

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/horizoncd/horizon/pkg/core/common"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	applicationdao "github.com/horizoncd/horizon/pkg/application/dao"
	appmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	envmanager "github.com/horizoncd/horizon/pkg/environment/manager"
	envmodels "github.com/horizoncd/horizon/pkg/environment/models"
	envregionmanager "github.com/horizoncd/horizon/pkg/environmentregion/manager"
	envregionmodels "github.com/horizoncd/horizon/pkg/environmentregion/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/group/models"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	regionmanager "github.com/horizoncd/horizon/pkg/region/manager"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	"github.com/horizoncd/horizon/pkg/server/global"
	tagmanager "github.com/horizoncd/horizon/pkg/tag/manager"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	// use tmp sqlite
	db, _        = orm.NewSqliteDB("")
	ctx          context.Context
	notExistID   = uint(100)
	Mgr          = New(db)
	regionMgr    = regionmanager.New(db)
	envMgr       = envmanager.New(db)
	envregionMgr = envregionmanager.New(db)
	tagMgr       = tagmanager.New(db)
)

func TestUint(t *testing.T) {
	g := models.Group{
		ParentID: 0,
	}

	_, err := json.Marshal(g)
	assert.Nil(t, err)
}

func getGroup(parentID uint, name, path string) *models.Group {
	return &models.Group{
		Name:            name,
		Path:            path,
		VisibilityLevel: "private",
		ParentID:        parentID,
		CreatedBy:       1,
		UpdatedBy:       1,
	}
}

func init() {
	// nolint
	ctx = context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name: "tony",
		ID:   110,
	})
	callbacks.RegisterCustomCallbacks(db)
	// create table
	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&appmodels.Application{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&membermodels.Member{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&envregionmodels.EnvironmentRegion{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&regionmodels.Region{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&envmodels.Environment{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
	err = db.AutoMigrate(&tagmodels.Tag{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}

func TestCreate(t *testing.T) {
	// normal create, parentID is nil
	g1, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	get, _ := Mgr.GetByID(ctx, g1.ID)
	assert.Equal(t, fmt.Sprintf("%d", g1.ID), get.TraversalIDs)

	// name conflict, parentID is nil
	_, err = Mgr.Create(ctx, getGroup(0, "1", "b"))
	assert.Equal(t, herrors.ErrNameConflict, perror.Cause(err))

	// path conflict, with parentID is nil
	_, err = Mgr.Create(ctx, getGroup(0, "2", "a"))
	assert.Equal(t, herrors.ErrPathConflict, perror.Cause(err))

	// name conflict with application
	name := "app"
	_, err = applicationdao.NewDAO(db).Create(ctx, &appmodels.Application{
		Name: name,
	}, nil)
	assert.Nil(t, err)
	_, err = Mgr.Create(ctx, getGroup(0, name, "a"))
	assert.Equal(t, perror.Cause(err), herrors.ErrGroupConflictWithApplication)

	// normal create, parentID: not nil
	group2 := getGroup(g1.ID, "2", "b")
	g2, err := Mgr.Create(ctx, group2)
	assert.Nil(t, err)
	get, _ = Mgr.GetByID(ctx, g2.ID)
	assert.Equal(t, fmt.Sprintf("%d,%d", g1.ID, g2.ID), get.TraversalIDs)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestDelete(t *testing.T) {
	g1, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)

	// delete exist record
	_, err = Mgr.Delete(ctx, g1.ID)
	assert.Nil(t, err)

	_, err = Mgr.GetByID(ctx, g1.ID)
	_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)

	// delete not exist record
	var count int64
	count, err = Mgr.Delete(ctx, notExistID)
	assert.Equal(t, 0, int(count))
	assert.Nil(t, err)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestGetByID(t *testing.T) {
	g1, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)

	// query exist record
	group1, err := Mgr.GetByID(ctx, g1.ID)
	assert.Nil(t, err)
	assert.NotNil(t, group1.ID)

	// query not exist record
	_, err = Mgr.GetByID(ctx, notExistID)
	_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestGetByIDs(t *testing.T) {
	g1, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	g2, _ := Mgr.Create(ctx, getGroup(0, "2", "b"))

	groups, err := Mgr.GetByIDs(ctx, []uint{g1.ID, g2.ID})
	assert.Nil(t, err)
	assert.Equal(t, g1.ID, groups[0].ID)
	assert.Equal(t, g2.ID, groups[1].ID)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestGetAll(t *testing.T) {
	g1, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	g2, err := Mgr.Create(ctx, getGroup(0, "2", "b"))
	assert.Nil(t, err)

	groups, err := Mgr.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, g1.ID, groups[0].ID)
	assert.Equal(t, g2.ID, groups[1].ID)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestGetByPaths(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	id2, _ := Mgr.Create(ctx, getGroup(0, "2", "b"))

	groups, err := Mgr.GetByPaths(ctx, []string{"a", "b"})
	assert.Nil(t, err)
	assert.Equal(t, id.ID, groups[0].ID)
	assert.Equal(t, id2.ID, groups[1].ID)

	// test GetByNameOrPathUnderParent
	groups, err = Mgr.GetByNameOrPathUnderParent(ctx, "1", "b", 0)
	assert.Nil(t, err)
	assert.Equal(t, len(groups), 2)
	assert.Equal(t, groups[0].Path, "a")
	assert.Equal(t, groups[1].Name, "2")

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestGetByNameFuzzily(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	id2, _ := Mgr.Create(ctx, getGroup(0, "21", "b"))

	groups, err := Mgr.GetByNameFuzzily(ctx, "1")
	assert.Nil(t, err)
	assert.Equal(t, id.ID, groups[0].ID)
	assert.Equal(t, id2.ID, groups[1].ID)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestUpdateBasic(t *testing.T) {
	group1 := getGroup(0, "1", "a")
	g1, err := Mgr.Create(ctx, group1)
	assert.Nil(t, err)

	// update exist record
	group1.ID = g1.ID
	group1.Name = "update1"
	err = Mgr.UpdateBasic(ctx, group1)
	assert.Nil(t, err)
	group, err := Mgr.GetByID(ctx, g1.ID)
	assert.Nil(t, err)
	assert.Equal(t, "update1", group.Name)

	// update fail because of conflict
	group2 := getGroup(0, "2", "b")
	g2, err := Mgr.Create(ctx, group2)
	assert.Nil(t, err)
	group2.ID = g2.ID
	group2.Name = "update1"
	err = Mgr.UpdateBasic(ctx, group2)
	assert.Equal(t, herrors.ErrNameConflict, perror.Cause(err))

	// update regionSelector
	err = Mgr.UpdateRegionSelector(ctx, g1.ID, "XXX")
	assert.Nil(t, err)
	group, _ = Mgr.GetByID(ctx, g1.ID)
	assert.Equal(t, group.RegionSelector, "XXX")

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestTransferGroup(t *testing.T) {
	g1, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	g2, err := Mgr.Create(ctx, getGroup(g1.ID, "2", "b"))
	assert.Nil(t, err)
	g3, err := Mgr.Create(ctx, getGroup(0, "3", "c"))
	assert.Nil(t, err)
	_, err = Mgr.Create(ctx, getGroup(g3.ID, "2", "d"))
	assert.Nil(t, err)

	// not valid transfer: name conflict
	err = Mgr.Transfer(ctx, g2.ID, g3.ID)
	assert.True(t, perror.Cause(err) == herrors.ErrNameConflict)

	// valid transfer
	err = Mgr.Transfer(ctx, g1.ID, g3.ID)
	assert.Nil(t, err)

	group, err := Mgr.GetByID(ctx, g2.ID)
	assert.Nil(t, err)

	expect := []string{
		strconv.Itoa(int(g3.ID)),
		strconv.Itoa(int(g1.ID)),
		strconv.Itoa(int(g2.ID)),
	}
	assert.Equal(t, strings.Join(expect, ","), group.TraversalIDs)
}

func TestManagerGetChildren(t *testing.T) {
	g0, err := Mgr.Create(ctx, getGroup(0, "0", "0"))
	assert.Nil(t, err)
	g1, err := Mgr.Create(ctx, getGroup(g0.ID, "1", "a"))
	assert.Nil(t, err)
	g2, err := Mgr.Create(ctx, getGroup(g0.ID, "2", "b"))
	assert.Nil(t, err)
	a1, err := applicationdao.NewDAO(db).Create(ctx, &appmodels.Application{
		Name:    "3",
		GroupID: g0.ID,
	}, nil)
	assert.Nil(t, err)

	type args struct {
		parentID   uint
		pageNumber int
		pageSize   int
	}
	tests := []struct {
		name    string
		args    args
		want    []*models.GroupOrApplication
		want1   int64
		wantErr bool
	}{
		{
			name: "firstPage",
			args: args{
				parentID:   g0.ID,
				pageNumber: 1,
				pageSize:   2,
			},
			want: []*models.GroupOrApplication{
				{
					Model: global.Model{
						ID:        g2.ID,
						UpdatedAt: g2.UpdatedAt,
					},
					Name:        "2",
					Path:        "b",
					Type:        "group",
					Description: "",
				},
				{
					Model: global.Model{
						ID:        g1.ID,
						UpdatedAt: g1.UpdatedAt,
					},
					Name:        "1",
					Path:        "a",
					Type:        "group",
					Description: "",
				},
			},
			want1: 3,
		},
		{
			name: "secondPage",
			args: args{
				parentID:   g0.ID,
				pageNumber: 2,
				pageSize:   2,
			},
			want: []*models.GroupOrApplication{
				{
					Model: global.Model{
						ID:        a1.ID,
						UpdatedAt: a1.UpdatedAt,
					},
					Name:        "3",
					Path:        "3",
					Type:        "application",
					Description: "",
				},
			},
			want1: 3,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, err := Mgr.GetChildren(ctx, tt.args.parentID, tt.args.pageNumber, tt.args.pageSize)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetChildren() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			for i, val := range got {
				val.UpdatedAt = tt.want[i].UpdatedAt
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetChildren() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("GetChildren() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestGetSubGroupsByGroupIDs(t *testing.T) {
	g1, err := Mgr.Create(ctx, getGroup(0, "a", "a"))
	assert.Nil(t, err)
	get, _ := Mgr.GetByID(ctx, g1.ID)
	assert.Equal(t, fmt.Sprintf("%d", g1.ID), get.TraversalIDs)

	g2, err := Mgr.Create(ctx, getGroup(0, "b", "b"))
	assert.Nil(t, err)
	get2, _ := Mgr.GetByID(ctx, g2.ID)
	assert.Equal(t, fmt.Sprintf("%d", g2.ID), get2.TraversalIDs)

	g3, err := Mgr.Create(ctx, getGroup(g1.ID, "c", "c"))
	assert.Nil(t, err)
	get3, _ := Mgr.GetByID(ctx, g3.ID)
	assert.Equal(t, fmt.Sprintf("%d,%d", g1.ID, g3.ID), get3.TraversalIDs)

	g4, err := Mgr.Create(ctx, getGroup(g2.ID, "c", "c"))
	assert.Nil(t, err)
	get4, _ := Mgr.GetByID(ctx, g4.ID)
	assert.Equal(t, fmt.Sprintf("%d,%d", g2.ID, g4.ID), get4.TraversalIDs)

	ids := []uint{g1.ID, g2.ID}
	groups, err := Mgr.GetSubGroupsByGroupIDs(ctx, ids)
	assert.Nil(t, err)
	assert.Equal(t, 4, len(groups))
	for _, group := range groups {
		t.Logf("group: %v", group)
	}

	ids2 := []uint{g2.ID}
	groups2, err := Mgr.GetSubGroupsByGroupIDs(ctx, ids2)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(groups2))
	for _, group := range groups2 {
		t.Logf("group: %v", group)
	}

	ids3 := []uint{g3.ID}
	groups3, err := Mgr.GetSubGroupsByGroupIDs(ctx, ids3)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(groups3))
	assert.Equal(t, g3.ID, groups3[0].ID)
	for _, group := range groups3 {
		t.Logf("group: %v", group)
	}
}

func Test_manager_GetSelectableRegionsByEnv(t *testing.T) {
	// initializing data
	r1, _ := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	_, _ = regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz-disabled",
		DisplayName: "HZ",
		Disabled:    true,
	})
	r3, _ := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz3",
		DisplayName: "HZ",
	})
	devEnv, _ := envMgr.CreateEnvironment(ctx, &envmodels.Environment{
		Name:        "dev",
		DisplayName: "开发",
	})
	_, _ = envregionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz",
		IsDefault:       true,
	})
	_, _ = envregionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz-disabled",
	})
	_, _ = envregionMgr.CreateEnvironmentRegion(ctx, &envregionmodels.EnvironmentRegion{
		EnvironmentName: devEnv.Name,
		RegionName:      "hz3",
	})
	_ = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r1.ID, []*tagmodels.Tag{
		{
			ResourceType: common.ResourceRegion,
			ResourceID:   r1.ID,
			Key:          "a",
			Value:        "1",
		}, {
			ResourceType: common.ResourceRegion,
			ResourceID:   r1.ID,
			Key:          "b",
			Value:        "1",
		},
	})
	_ = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r3.ID, []*tagmodels.Tag{
		{
			ResourceType: common.ResourceRegion,
			ResourceID:   r3.ID,
			Key:          "a",
			Value:        "1",
		}, {
			ResourceType: common.ResourceRegion,
			ResourceID:   r3.ID,
			Key:          "c",
			Value:        "1",
		},
	})
	g1, err := Mgr.Create(ctx, &models.Group{
		Name: "11",
		Path: "pp",
		RegionSelector: `- key: "a"
  operator: "in"
  values: 
    - "1"
- key: "b"
  operator: "in"
  values: 
    - "1"
`,
	})
	assert.Nil(t, err)
	// get regionSelector from parent group
	g2, _ := Mgr.Create(ctx, &models.Group{
		Name:     "22",
		Path:     "p2",
		ParentID: g1.ID,
	})

	type args struct {
		id  uint
		env string
	}
	tests := []struct {
		name    string
		args    args
		want    regionmodels.RegionParts
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				id:  g2.ID,
				env: "dev",
			},
			want: regionmodels.RegionParts{
				{
					Name:        "hz",
					DisplayName: "HZ",
					IsDefault:   true,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mgr.GetSelectableRegionsByEnv(ctx, tt.args.id, tt.args.env)
			if (err != nil) != tt.wantErr {
				t.Errorf(fmt.Sprintf("GetSelectableRegionsByEnv(%v, %v, %v)", ctx, tt.args.id, tt.args.env))
			}
			assert.Equalf(t, tt.want, got, "GetSelectableRegionsByEnv(%v, %v, %v)", ctx, tt.args.id, tt.args.env)
		})
	}

	defaultRegions, err := Mgr.GetDefaultRegions(ctx, g2.ID)
	assert.Nil(t, err)
	assert.Equal(t, len(defaultRegions), 1)
	assert.Equal(t, defaultRegions[0].RegionName, "hz")
	assert.Equal(t, defaultRegions[0].EnvironmentName, "dev")
}

func Test_manager_GetSelectableRegions(t *testing.T) {
	r1, _ := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz",
		DisplayName: "HZ",
	})
	_, _ = regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz-disabled",
		DisplayName: "HZ",
		Disabled:    true,
	})
	r3, _ := regionMgr.Create(ctx, &regionmodels.Region{
		Name:        "hz3",
		DisplayName: "HZ",
	})

	_ = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r1.ID, []*tagmodels.Tag{
		{
			ResourceType: common.ResourceRegion,
			ResourceID:   r1.ID,
			Key:          "a",
			Value:        "11",
		},
	})
	_ = tagMgr.UpsertByResourceTypeID(ctx, common.ResourceRegion, r3.ID, []*tagmodels.Tag{
		{
			ResourceType: common.ResourceRegion,
			ResourceID:   r3.ID,
			Key:          "a",
			Value:        "11",
		},
	})

	g1, err := Mgr.Create(ctx, &models.Group{
		Name: "112",
		Path: "pp2",
		RegionSelector: `- key: "a"
  operator: "in"
  values: 
    - "11"
`,
	})
	assert.Nil(t, err)
	type args struct {
		id uint
	}
	tests := []struct {
		name    string
		args    args
		want    regionmodels.RegionParts
		wantErr bool
	}{
		{
			name: "normal",
			args: args{
				id: g1.ID,
			},
			want: regionmodels.RegionParts{
				{
					Name:        "hz",
					DisplayName: "HZ",
				},
				{
					Name:        "hz3",
					DisplayName: "HZ",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Mgr.GetSelectableRegions(ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf(fmt.Sprintf("GetSelectableRegions(%v, %v)", ctx, tt.args.id))
			}
			assert.Equalf(t, tt.want, got, "GetSelectableRegions(%v, %v)", tt.args.id)
		})
	}
}
