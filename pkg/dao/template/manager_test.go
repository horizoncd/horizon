package template

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/pkg/lib/orm"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	var (
		name        = "javaapp"
		description = "javaapp template"
		createdBy   = "tony"
		updatedBy   = "tony"
	)
	template := &Template{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
	}
	template, err := Mgr.Create(ctx, template)
	assert.Nil(t, err)

	assert.Equal(t, name, template.Name)
	assert.Equal(t, description, template.Description)
	assert.Equal(t, 1, int(template.ID))

	b, err := json.Marshal(template)
	assert.Nil(t, err)
	t.Logf(string(b))

	templates, err := Mgr.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, name, templates[0].Name)
	assert.Equal(t, description, templates[0].Description)
	assert.Equal(t, 1, int(templates[0].ID))
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&Template{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
