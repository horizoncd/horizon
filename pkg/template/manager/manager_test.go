package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/template/models"
	"github.com/stretchr/testify/assert"
)

var (
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func Test(t *testing.T) {
	var (
		name        = "javaapp"
		description = "javaapp template"
		createdBy   = uint(1)
		updatedBy   = uint(1)
	)
	template := &models.Template{
		Name:        name,
		Description: description,
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
	}
	template, err := mgr.Create(ctx, template)
	assert.Nil(t, err)

	assert.Equal(t, name, template.Name)
	assert.Equal(t, description, template.Description)
	assert.Equal(t, 1, int(template.ID))

	b, err := json.Marshal(template)
	assert.Nil(t, err)
	t.Logf(string(b))

	templates, err := mgr.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, name, templates[0].Name)
	assert.Equal(t, description, templates[0].Description)
	assert.Equal(t, 1, int(templates[0].ID))
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.Template{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
