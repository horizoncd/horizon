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

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	applicationdao "g.hz.netease.com/horizon/pkg/application/dao"
	appmodels "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   = orm.NewContext(context.TODO(), db)

	notExistID = uint(100)
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
}

func TestCreate(t *testing.T) {
	// normal create, parentID is nil
	g1, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	get, _ := Mgr.GetByID(ctx, g1.ID)
	assert.Equal(t, fmt.Sprintf("%d", g1.ID), get.TraversalIDs)

	// name conflict, parentID is nil
	_, err = Mgr.Create(ctx, getGroup(0, "1", "b"))
	assert.Equal(t, common.ErrNameConflict, err)

	// path conflict, with parentID is nil
	_, err = Mgr.Create(ctx, getGroup(0, "2", "a"))
	assert.Equal(t, dao.ErrPathConflict, err)

	// name conflict with application
	name := "app"
	_, err = applicationdao.NewDAO().Create(ctx, &appmodels.Application{
		Name: name,
	})
	assert.Nil(t, err)
	_, err = Mgr.Create(ctx, getGroup(0, name, "a"))
	assert.Equal(t, err, ErrConflictWithApplication)

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
	assert.Equal(t, err, gorm.ErrRecordNotFound)

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
	assert.Equal(t, err, gorm.ErrRecordNotFound)

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

func TestGetByPaths(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	id2, _ := Mgr.Create(ctx, getGroup(0, "2", "b"))

	groups, err := Mgr.GetByPaths(ctx, []string{"a", "b"})
	assert.Nil(t, err)
	assert.Equal(t, id.ID, groups[0].ID)
	assert.Equal(t, id2.ID, groups[1].ID)

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
	assert.Equal(t, common.ErrNameConflict, err)

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

	err = Mgr.Transfer(ctx, g1.ID, g3.ID, 1)
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
	a1, err := applicationdao.NewDAO().Create(ctx, &appmodels.Application{
		Name:    "3",
		GroupID: g0.ID,
	})
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
					Model: gorm.Model{
						ID:        g2.ID,
						UpdatedAt: g2.UpdatedAt,
					},
					Name:        "2",
					Path:        "b",
					Type:        "group",
					Description: "",
				},
				{
					Model: gorm.Model{
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
					Model: gorm.Model{
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
