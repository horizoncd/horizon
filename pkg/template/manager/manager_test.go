package manager_test

import (
	"context"
	"os"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	applicationmodel "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	clustermodel "github.com/horizoncd/horizon/pkg/cluster/models"
	"github.com/horizoncd/horizon/pkg/core/common"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/global"
	"github.com/horizoncd/horizon/pkg/template/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	ctx    context.Context
	params *managerparam.Manager
)

func Test(t *testing.T) {
	template1 := &models.Template{
		Name:        "javaapp1",
		Description: "java app for test",
		GroupID:     1,
	}
	template1InDB, err := params.TemplateMgr.Create(ctx, template1)
	assert.Nil(t, err)

	assert.Equal(t, template1.Name, template1InDB.Name)
	assert.Equal(t, template1.Description, template1InDB.Description)
	assert.Equal(t, 1, int(template1.ID))

	template2 := &models.Template{
		Name:        "javaapp2",
		Description: "java app for test 2",
		GroupID:     2,
	}

	_, err = params.TemplateMgr.Create(ctx, template2)
	assert.Nil(t, err)

	template2InDB, err := params.TemplateMgr.GetByID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, template2.Name, template2InDB.Name)
	assert.Equal(t, template2.Description, template2InDB.Description)
	assert.Equal(t, template2.WithoutCI, false)

	template2InDB, err = params.TemplateMgr.GetByName(ctx, template2.Name)
	assert.Nil(t, err)
	assert.Equal(t, template2.Description, template2InDB.Description)

	templates, err := params.TemplateMgr.ListTemplate(ctx)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(templates))
	assert.Equal(t, template1.Name, templates[0].Name)
	assert.Equal(t, template1.Description, templates[0].Description)
	assert.Equal(t, 1, int(templates[0].ID))
	assert.Equal(t, template2.Name, templates[1].Name)
	assert.Equal(t, template2.Description, templates[1].Description)
	assert.Equal(t, 2, int(templates[1].ID))

	templates, err = params.TemplateMgr.ListByGroupID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, template2.Name, templates[0].Name)
	assert.Equal(t, template2.Description, templates[0].Description)
	assert.Equal(t, 2, int(templates[0].ID))

	template2.Description = "changed description"
	err = params.TemplateMgr.UpdateByID(ctx, 2, template2)
	assert.Nil(t, err)

	template2InDB, err = params.TemplateMgr.GetByID(ctx, 2)
	assert.Nil(t, err)
	assert.Equal(t, template2.Name, template2InDB.Name)
	assert.Equal(t, template2.Description, template2InDB.Description)

	templates, err = params.TemplateMgr.ListByGroupIDs(ctx, []uint{1, 2})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 2, len(templates))

	templates, err = params.TemplateMgr.ListByGroupIDs(ctx, []uint{1})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, template1.Name, templates[0].Name)

	templates, err = params.TemplateMgr.ListByIDs(ctx, []uint{1, 2})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 2, len(templates))

	templates, err = params.TemplateMgr.ListByIDs(ctx, []uint{1})
	assert.Nil(t, err)
	assert.NotNil(t, templates)
	assert.Equal(t, 1, len(templates))
	assert.Equal(t, template1.Name, templates[0].Name)

	err = params.TemplateMgr.DeleteByID(ctx, 2)
	assert.Nil(t, err)

	template2InDB, err = params.TemplateMgr.GetByID(ctx, 2)
	assert.NotNil(t, err)
	assert.Nil(t, template2InDB)

	app := &applicationmodel.Application{
		Model:    global.Model{ID: 1},
		Template: template1.Name,
		Name:     "test",
	}
	_, err = params.ApplicationManager.Create(ctx, app, map[string]string{})
	assert.Nil(t, err)

	cluster := &clustermodel.Cluster{
		Model:         global.Model{ID: 1},
		ApplicationID: 1,
		Name:          "testgroup",
		Template:      template1.Name,
	}
	_, err = params.ClusterMgr.Create(ctx, cluster, nil, nil)
	assert.Nil(t, err)

	apps, _, err := params.TemplateMgr.GetRefOfApplication(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, app.Name, apps[0].Name)

	clusters, _, err := params.TemplateMgr.GetRefOfCluster(ctx, 1)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, cluster.Name, clusters[0].Name)
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

	params = managerparam.InitManager(db)

	os.Exit(m.Run())
}
