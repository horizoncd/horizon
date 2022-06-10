package manager

import (
	"context"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/harbor/models"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	"github.com/stretchr/testify/assert"
)

var (
	// use tmp sqlite
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func init() {
	if err := db.AutoMigrate(&models.Harbor{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&regionmodels.Region{}); err != nil {
		panic(err)
	}

	ctx = context.TODO()
}

func Test(t *testing.T) {
	id, err := mgr.Create(ctx, &models.Harbor{
		Name:            "1",
		Server:          "2",
		Token:           "1",
		PreheatPolicyID: 0,
	})
	assert.Nil(t, err)

	harbor, err := mgr.GetByID(ctx, id)
	assert.Nil(t, err)
	assert.Equal(t, harbor.Name, "1")
	assert.Equal(t, harbor.Server, "2")
	assert.Equal(t, harbor.Token, "1")
	assert.Equal(t, harbor.PreheatPolicyID, 0)

	err = mgr.UpdateByID(ctx, id, &models.Harbor{
		Name:            "2",
		Server:          "1",
		Token:           "2",
		PreheatPolicyID: 1,
	})
	assert.Nil(t, err)
	harbor, _ = mgr.GetByID(ctx, id)
	assert.Equal(t, harbor.Name, "2")
	assert.Equal(t, harbor.Server, "1")
	assert.Equal(t, harbor.Token, "2")
	assert.Equal(t, harbor.PreheatPolicyID, 1)

	err = mgr.DeleteByID(ctx, id)
	assert.Nil(t, err)
	harbor, err = mgr.GetByID(ctx, id)
	assert.NotNil(t, err)
	assert.Nil(t, harbor)
}
