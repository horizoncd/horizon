package manager

import (
	"context"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	amanager "g.hz.netease.com/horizon/pkg/application/manager"
	applicationmodel "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermodel "g.hz.netease.com/horizon/pkg/cluster/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/template/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db             *gorm.DB
	ctx            context.Context
	templateMgr    Manager
	applicationMgr amanager.Manager
)

func Test(t *testing.T) {
	template1 := &models.Template{
		Name:        "javaapp1",
		Description: "java app for test",
		GroupID:     1,
	}
	template1InDB, err := templateMgr.Create(ctx, template1)
	assert.Nil(t, err)

	assert.Equal(t, template1.Name, template1InDB.Name)
	assert.Equal(t, template1.Description, template1InDB.Description)
	assert.Equal(t, 1, int(template1.ID))

	template2 := &models.Template{
		Name:        "javaapp2",
		Description: "java app for test 2",
		GroupID:     2,
	}

	_, err = templateMgr.Create(ctx, template2)
	assert.Nil(t, err)

	template2InDB, err := templateMgr.GetByID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, template2.Name, template2InDB.Name)
	assert.Equal(t, template2.Description, template2InDB.Description)

	template2InDB, err = templateMgr.GetByName(ctx, template2.Name)
	assert.Nil(t, err)
	assert.Equal(t, template2.Description, template2InDB.Description)

	templates, err := templateMgr.List(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))
	assert.Equal(t, template1.Name, templates[0].Name)
	assert.Equal(t, template1.Description, templates[0].Description)
	assert.Equal(t, 1, int(templates[0].ID))
	assert.Equal(t, template2.Name, templates[1].Name)
	assert.Equal(t, template2.Description, templates[1].Description)
	assert.Equal(t, 2, int(templates[1].ID))

	templates, err = templateMgr.ListByGroupID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, template2.Name, templates[0].Name)
	assert.Equal(t, template2.Description, templates[0].Description)
	assert.Equal(t, 2, int(templates[0].ID))

	template2.Description = "changed description"
	err = templateMgr.UpdateByID(ctx, 2, template2)
	assert.Nil(t, err)

	template2InDB, err = templateMgr.GetByID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, template2.Name, template2InDB.Name)
	assert.Equal(t, template2.Description, template2InDB.Description)

	err = templateMgr.DeleteByID(ctx, 2)
	assert.Nil(t, err)

	template2InDB, err = templateMgr.GetByID(ctx, 2)
	assert.NotNil(t, err)
	assert.Nil(t, template2InDB)

	app := &applicationmodel.Application{
		Template: template1.Name,
		Name:     "test",
	}
	_, err = applicationMgr.Create(ctx, app, map[string]string{})
	assert.Nil(t, err)

	apps, _, err := templateMgr.GetRefOfApplication(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, app.Name, apps[0].Name)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Template{}, &trmodels.TemplateRelease{},
		&membermodels.Member{},
		&applicationmodel.Application{}, &clustermodel.Cluster{}); err != nil {
		panic(err)
	}
	ctx = context.Background()
	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		ID:   1,
		Name: "Jerry",
	})

	applicationMgr = amanager.New(db)
	templateMgr = New(db)

	os.Exit(m.Run())
}
