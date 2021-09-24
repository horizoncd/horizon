package group

import (
	"context"
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
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

func getGroup(parentID int, name, path string) *models.Group {
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

	// name conflict, parentId: nil
	_, err = Mgr.Create(ctx, getGroup(0, "1", "b"))
	assert.Equal(t, common.ErrNameConflict, err)

	// path conflict, parentId: nil
	_, err = Mgr.Create(ctx, getGroup(0, "2", "a"))
	assert.Equal(t, ErrPathConflict, err)

	// normal create, parent: 1
	group2 := getGroup(int(id), "2", "b")
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
	err = Mgr.Delete(ctx, id)
	assert.Nil(t, err)

	_, err = Mgr.GetByID(ctx, id)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// delete not exist record
	err = Mgr.Delete(ctx, notExistID)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

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

	// update not exist record
	group1.ID = notExistID
	group1.Name = "update2"
	err = Mgr.UpdateBasic(ctx, group1)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// drop table
	res := db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&models.Group{})
	assert.Nil(t, res.Error)
}

func TestList(t *testing.T) {
	pid, err := Mgr.Create(ctx, getGroup(0, "1", "a"))
	assert.Nil(t, err)
	var group2Id, group3Id uint
	group2Id, err = Mgr.Create(ctx, getGroup(int(pid), "2", "b"))
	assert.Nil(t, err)
	group3Id, err = Mgr.Create(ctx, getGroup(int(pid), "3", "c"))
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
	id2, err := Mgr.Create(ctx, getGroup(int(id), "2", "b"))
	assert.Nil(t, err)
	id3, err := Mgr.Create(ctx, getGroup(0, "3", "c"))
	assert.Nil(t, err)

	err = Mgr.Transfer(ctx, id, id3)
	assert.Nil(t, err)

	group, err := Mgr.GetByID(ctx, id2)
	assert.Nil(t, err)

	assert.Equal(t, "3,1,2", group.TraversalIDs)
}
