package group

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/dao"
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
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
	}
}

func init() {
	// create table
	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		os.Exit(1)
	}
}

func TestCreate(t *testing.T) {
	// normal create
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	get, _ := Mgr.GetByID(ctx, id)
	assert.Equal(t, fmt.Sprintf("%d", id), get.TraversalIDs)

	// name conflict, parentID: nil
	_, err = Mgr.Create(ctx, getGroup(0, "1", "b"))
	assert.Equal(t, common.ErrNameConflict, err)

	// path conflict, parentID: nil
	_, err = Mgr.Create(ctx, getGroup(0, "2", "a"))
	assert.Equal(t, dao.ErrPathConflict, err)

	// normal create, parent: 1
	group2 := getGroup(id, "2", "b")
	id2, err := Mgr.Create(ctx, group2)
	assert.Nil(t, err)
	get, _ = Mgr.GetByID(ctx, id2)
	assert.Equal(t, fmt.Sprintf("%d,%d", id, id2), get.TraversalIDs)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestDelete(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)

	// delete exist record
	_, err = Mgr.Delete(ctx, id)
	assert.Nil(t, err)

	_, err = Mgr.GetByID(ctx, id)
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
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)

	// query exist record
	group1, err := Mgr.GetByID(ctx, id)
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
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	id2, _ := Mgr.Create(ctx, getGroup(0, "2", "b"))

	groups, err := Mgr.GetByIDs(ctx, []uint{id, id2})
	assert.Nil(t, err)
	assert.Equal(t, id, groups[0].ID)
	assert.Equal(t, id2, groups[1].ID)

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
	assert.Equal(t, id, groups[0].ID)
	assert.Equal(t, id2, groups[1].ID)

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
	assert.Equal(t, id, groups[0].ID)
	assert.Equal(t, id2, groups[1].ID)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestGetByIDsOrderByIDDesc(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	id2, _ := Mgr.Create(ctx, getGroup(0, "2", "b"))

	groups, err := Mgr.GetByIDsOrderByIDDesc(ctx, []uint{id, id2})
	assert.Nil(t, err)
	assert.Equal(t, id, groups[1].ID)
	assert.Equal(t, id2, groups[0].ID)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestUpdateBasic(t *testing.T) {
	group1 := getGroup(0, "1", "a")
	id, err := Mgr.Create(ctx, group1)
	assert.Nil(t, err)

	// update exist record
	group1.ID = id
	group1.Name = "update1"
	err = Mgr.UpdateBasic(ctx, group1)
	assert.Nil(t, err)
	group, err := Mgr.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, "update1", group.Name)

	// update fail because of conflict
	group2 := getGroup(0, "2", "b")
	id2, err := Mgr.Create(ctx, group2)
	assert.Nil(t, err)
	group2.ID = id2
	group2.Name = "update1"
	err = Mgr.UpdateBasic(ctx, group2)
	assert.Equal(t, common.ErrNameConflict, err)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestList(t *testing.T) {
	pid, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	var group2Id, group3Id uint
	group2Id, err = Mgr.Create(ctx, getGroup(pid, "2", "b"))
	assert.Nil(t, err)
	group3Id, err = Mgr.Create(ctx, getGroup(pid, "3", "c"))
	assert.Nil(t, err)

	// page with keywords, items: 1, total: 1
	query := q.New(q.KeyWords{
		"name": "2",
	})
	query.PageNumber = 1
	query.PageSize = 1
	items, total, err := Mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, int64(1), total)

	// page without keywords, items: 1, total: 2
	query.Keywords = q.KeyWords{}
	items, total, err = Mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(items))
	assert.Equal(t, int64(3), total)

	// list by parentIdList
	query.Keywords = q.KeyWords{
		"parent_id": []uint{
			pid,
		},
	}
	query.PageSize = 10
	items, _, err = Mgr.List(ctx, query)
	assert.Nil(t, err)
	assert.Equal(t, group2Id, items[0].ID)
	assert.Equal(t, group3Id, items[1].ID)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestTransferGroup(t *testing.T) {
	id, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	id2, err := Mgr.Create(ctx, getGroup(id, "2", "b"))
	assert.Nil(t, err)
	id3, err := Mgr.Create(ctx, getGroup(0, "3", "c"))
	assert.Nil(t, err)

	err = Mgr.Transfer(ctx, id, id3)
	assert.Nil(t, err)

	group, err := Mgr.GetByID(ctx, id2)
	assert.Nil(t, err)

	expect := []string{
		strconv.Itoa(int(id3)),
		strconv.Itoa(int(id)),
		strconv.Itoa(int(id2)),
	}
	assert.Equal(t, strings.Join(expect, ","), group.TraversalIDs)
}
