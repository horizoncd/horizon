package templaterelease

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	var (
		templateName  = "javaapp"
		name          = "v1.0.0"
		description   = "javaapp template v1.0.0"
		gitlabName    = "control"
		gitlabProject = "helm-template/javaapp"
		createdBy     = "tony"
		updatedBy     = "tony"
	)
	templateRelease := &models.TemplateRelease{
		TemplateName:  templateName,
		Name:          name,
		Description:   description,
		GitlabName:    gitlabName,
		GitlabProject: gitlabProject,
		Recommended:   true,
		CreatedBy:     createdBy,
		UpdatedBy:     updatedBy,
	}
	templateRelease, err := Mgr.Create(ctx, templateRelease)
	assert.Nil(t, err)

	assert.Equal(t, name, templateRelease.Name)
	assert.Equal(t, description, templateRelease.Description)
	assert.Equal(t, 1, int(templateRelease.ID))

	b, err := json.Marshal(templateRelease)
	assert.Nil(t, err)
	t.Logf(string(b))

	templates, err := Mgr.ListByTemplateName(ctx, templateName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, name, templates[0].Name)
	assert.Equal(t, description, templates[0].Description)
	assert.Equal(t, 1, int(templates[0].ID))
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.TemplateRelease{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
