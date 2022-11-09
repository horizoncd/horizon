package manager

import (
	"context"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"g.hz.netease.com/horizon/pkg/registry/models"
	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func init() {
	if err := db.AutoMigrate(&models.Registry{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}

	ctx = context.TODO()
}

func Test(t *testing.T) {
	id, err := mgr.Create(ctx, &models.Registry{
		Name:   "1",
		Server: "2",
		Token:  "1",
	})
	assert.Nil(t, err)

	registry, err := mgr.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, registry.Name, "1")
	assert.Equal(t, registry.Server, "2")
	assert.Equal(t, registry.Token, "1")

	err = mgr.UpdateByID(ctx, id, &models.Registry{
		Name:   "2",
		Server: "1",
		Token:  "2",
	})
	assert.Nil(t, err)
	registry, _ = mgr.GetByID(ctx, id)
	assert.Equal(t, registry.Name, "2")
	assert.Equal(t, registry.Server, "1")
	assert.Equal(t, registry.Token, "2")

	err = mgr.DeleteByID(ctx, id)
	assert.Nil(t, err)
	registry, err = mgr.GetByID(ctx, id)
	assert.NotNil(t, err)
	assert.Nil(t, registry)
}
