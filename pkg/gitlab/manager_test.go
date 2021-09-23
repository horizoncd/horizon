package gitlab

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/gitlab/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	var (
		name      = "control"
		url       = "https://gitlab.com"
		token     = "token"
		createdBy = "tony"
		updatedBy = "tony"
	)
	gitlab := &models.Gitlab{
		Name:      name,
		URL:       url,
		Token:     token,
		CreatedBy: createdBy,
		UpdatedBy: updatedBy,
	}
	gitlab, err := Mgr.Create(ctx, gitlab)
	assert.Nil(t, err)

	assert.Equal(t, name, gitlab.Name)
	assert.Equal(t, 1, int(gitlab.ID))

	b, err := json.Marshal(gitlab)
	assert.Nil(t, err)
	t.Logf(string(b))

	gitlabs, err := Mgr.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(gitlabs))
	assert.Equal(t, name, gitlabs[0].Name)
	assert.Equal(t, 1, int(gitlabs[0].ID))

	// get by name
	gitlab, err = Mgr.GetByName(ctx, name)
	assert.Nil(t, err)
	assert.NotNil(t, gitlab)

	// get by name not exists
	gitlab, err = Mgr.GetByName(ctx, "not-exists")
	assert.Nil(t, err)
	assert.Nil(t, gitlab)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Gitlab{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
