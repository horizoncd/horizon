package group

import (
	"fmt"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/group/models"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   = orm.NewContext(nil, db)

	group1Id   = uint(1)
	group1Path = "/a"
	group1     = &models.Group{
		Name:            "1",
		Path:            "a",
		VisibilityLevel: "private",
	}

	pid    = uint(1)
	group2 = &models.Group{
		Name:            "2",
		Path:            "b",
		VisibilityLevel: "public",
		ParentId:        &pid,
	}

	notExistId   = uint(100)
	notExistPath = "x"
)

func init() {
	// create table
	err := db.AutoMigrate(&models.Group{})
	if err != nil {
		fmt.Printf("%+v", err)
		return
	}
	// create first record
	id, err := Mgr.Create(ctx, group1)
	if err != nil {
		fmt.Printf("%+v", err)
		return
	}
	if id != uint(1) {
		fmt.Print("create first record fail")
		os.Exit(1)
	}
}

func TestCreate(t *testing.T) {
	_, err := Mgr.Create(ctx, group2)
	assert.Nil(t, err)
	assert.Equal(t, "/a/b", group2.Path)
	assert.Equal(t, "1 / 2", group2.FullName)
}

func TestDelete(t *testing.T) {
	// delete exist record
	group1, err := Mgr.Get(ctx, group1Id)
	assert.Nil(t, err)
	assert.NotNil(t, group1)

	err = Mgr.Delete(ctx, group1Id)
	assert.Nil(t, err)

	_, err = Mgr.Get(ctx, group1Id)
	assert.Equal(t, err, gorm.ErrRecordNotFound)

	// delete not exist record
	err = Mgr.Delete(ctx, notExistId)
	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestGet(t *testing.T) {
	// query exist record
	group1, err := Mgr.Get(ctx, group1Id)
	assert.Nil(t, err)
	assert.NotNil(t, group1)

	// query not exist record
	_, err = Mgr.Get(ctx, notExistId)
	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestGetByPath(t *testing.T) {
	// query exist record
	group1, err := Mgr.GetByPath(ctx, group1Path)
	assert.Nil(t, err)
	assert.NotNil(t, group1)

	// query not exist record
	_, err = Mgr.GetByPath(ctx, notExistPath)
	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestUpdate(t *testing.T) {
	// update exist record
	group1.ID = group1Id
	group1.Name = "update1"
	err := Mgr.Update(ctx, group1)
	assert.Nil(t, err)

	// update not exist record
	group1.ID = notExistId
	err = Mgr.Update(ctx, group1)
	assert.Equal(t, err, gorm.ErrRecordNotFound)
}

func TestList(t *testing.T) {
	_, err := Mgr.Create(ctx, group2)

	// page with keywords, items: 1, total: 1
	query := q.New(q.KeyWords{
		"name": "1",
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
	assert.Equal(t, int64(2), total)
}
