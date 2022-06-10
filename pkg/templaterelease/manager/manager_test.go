package manager

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/util/errors"

	"github.com/stretchr/testify/assert"
)

var (
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func Test(t *testing.T) {
	var (
		templateName  = "javaapp"
		name          = "v1.0.0"
		description   = "javaapp template v1.0.0"
		gitlabProject = "helm-template/javaapp"
		createdBy     = uint(1)
		updatedBy     = uint(1)
	)
	templateRelease := &models.TemplateRelease{
		TemplateName:  templateName,
		Name:          name,
		Description:   description,
		GitlabProject: gitlabProject,
		Recommended:   true,
		CreatedBy:     createdBy,
		UpdatedBy:     updatedBy,
	}
	templateRelease, err := mgr.Create(ctx, templateRelease)
	assert.Nil(t, err)

	assert.Equal(t, name, templateRelease.Name)
	assert.Equal(t, description, templateRelease.Description)
	assert.Equal(t, 1, int(templateRelease.ID))

	b, err := json.Marshal(templateRelease)
	assert.Nil(t, err)
	t.Logf(string(b))

	templates, err := mgr.ListByTemplateName(ctx, templateName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, name, templates[0].Name)
	assert.Equal(t, description, templates[0].Description)
	assert.Equal(t, 1, int(templates[0].ID))

	template, err := mgr.GetByTemplateNameAndRelease(ctx, templateName, name)
	assert.Nil(t, err)
	assert.NotNil(t, template)
	assert.Equal(t, name, templateRelease.Name)

	// template release not exists
	template, err = mgr.GetByTemplateNameAndRelease(ctx, templateName, "not-exist")
	assert.NotNil(t, err)
	assert.Equal(t, http.StatusNotFound, errors.Status(err))
	assert.Nil(t, template)
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.TemplateRelease{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
