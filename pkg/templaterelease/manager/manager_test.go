package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	amanager "github.com/horizoncd/horizon/pkg/application/manager"
	applicationmodel "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/core/common"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	tmanager "github.com/horizoncd/horizon/pkg/template/manager"
	tmodels "github.com/horizoncd/horizon/pkg/template/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db                 *gorm.DB
	ctx                context.Context
	templateMgr        tmanager.Manager
	templateReleaseMgr Manager
	applicationMgr     amanager.Manager
)

func Test(t *testing.T) {
	var (
		templateName = "javaapp"
		name         = "v1.0.0"
		chartVersion = "v1.0.0-test"
		repo         = "repo"
		description  = "javaapp template v1.0.0"
		groupID      = uint(0)
		createdBy    = uint(1)
		updatedBy    = uint(1)
		err          error
	)
	template := &tmodels.Template{
		Name:        templateName,
		Description: description,
		Repository:  repo,
		GroupID:     groupID,
		CreatedBy:   createdBy,
		UpdatedBy:   updatedBy,
	}
	template, err = templateMgr.Create(ctx, template)
	assert.Nil(t, err)

	recommend := true

	templateRelease := &trmodels.TemplateRelease{
		Template:     template.ID,
		TemplateName: templateName,
		Name:         name,
		ChartVersion: chartVersion,
		Description:  description,
		Recommended:  &recommend,
		CreatedBy:    createdBy,
		UpdatedBy:    updatedBy,
	}
	templateRelease, err = templateReleaseMgr.Create(ctx, templateRelease)
	assert.Nil(t, err)

	assert.Equal(t, name, templateRelease.Name)
	assert.Equal(t, description, templateRelease.Description)
	assert.Equal(t, 1, int(templateRelease.ID))

	b, err := json.Marshal(templateRelease)
	assert.Nil(t, err)
	t.Logf(string(b))

	releases, err := templateReleaseMgr.ListByTemplateName(ctx, templateName)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, name, releases[0].Name)
	assert.Equal(t, chartVersion, releases[0].ChartVersion)
	assert.Equal(t, description, releases[0].Description)
	assert.Equal(t, 1, int(releases[0].ID))

	releases, err = templateReleaseMgr.ListByTemplateID(ctx, template.ID)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(releases))
	assert.Equal(t, name, releases[0].Name)
	assert.Equal(t, chartVersion, releases[0].ChartVersion)
	assert.Equal(t, description, releases[0].Description)
	assert.Equal(t, 1, int(releases[0].ID))

	// template release not exists
	templateRelease, err = templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, "not-exist")
	assert.NotNil(t, err)
	_, ok := perror.Cause(err).(*herrors.HorizonErrNotFound)
	assert.True(t, ok)
	assert.Nil(t, templateRelease)

	templateRelease, err = templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, name)
	assert.Nil(t, err)
	assert.NotNil(t, templateRelease)
	assert.Equal(t, chartVersion, templateRelease.ChartVersion)
	assert.Equal(t, name, templateRelease.Name)

	app := &applicationmodel.Application{
		Template:        templateName,
		TemplateRelease: templateRelease.Name,
		Name:            "test",
	}
	_, err = applicationMgr.Create(ctx, app, map[string]string{})
	assert.Nil(t, err)

	apps, _, err := templateReleaseMgr.GetRefOfApplication(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, app.Name, apps[0].Name)

	err = templateReleaseMgr.DeleteByID(ctx, templateRelease.ID)
	assert.Nil(t, err)

	templateRelease, err = templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, name)
	assert.NotNil(t, err)
	assert.Nil(t, templateRelease)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&trmodels.TemplateRelease{},
		&applicationmodel.Application{}, &tmodels.Template{},
		&membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		ID:   1,
		Name: "Jerry",
	})

	templateMgr = tmanager.New(db)
	templateReleaseMgr = New(db)
	applicationMgr = amanager.New(db)

	os.Exit(m.Run())
}
