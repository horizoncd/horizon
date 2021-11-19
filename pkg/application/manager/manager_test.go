package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/application/models"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	var (
		groupID         = 1
		name            = "application"
		description     = "description about application"
		priority        = models.P0
		gitURL          = "ssh://git@github.com"
		gitSubfolder    = "/"
		gitBranch       = "develop"
		template        = "javaapp"
		templateRelease = "v1.1.0"
		createdBy       = uint(1)
		updatedBy       = uint(1)
	)
	application := &models.Application{
		GroupID:         uint(groupID),
		Name:            name,
		Description:     description,
		Priority:        priority,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitBranch:       gitBranch,
		Template:        template,
		TemplateRelease: templateRelease,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
	}
	application, err := Mgr.Create(ctx, application)
	assert.Nil(t, err)

	assert.Equal(t, name, application.Name)
	assert.Equal(t, 1, int(application.ID))

	b, err := json.Marshal(application)
	assert.Nil(t, err)
	t.Logf(string(b))

	appGetByID, err := Mgr.GetByID(ctx, application.ID)
	assert.Nil(t, err)
	assert.Equal(t, application.Name, appGetByID.Name)

	appGetByName, err := Mgr.GetByName(ctx, application.Name)
	assert.Nil(t, err)
	assert.Equal(t, application.ID, appGetByName.ID)

	appGetByName.Description = "new"
	appGetByName, err = Mgr.UpdateByID(ctx, application.ID, appGetByName)
	assert.Nil(t, err)

	apps, err := Mgr.GetByNameFuzzily(ctx, "app")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, appGetByName.Name, apps[0].Name)

	total, apps, err := Mgr.GetByNameFuzzilyByPagination(ctx, "app", q.Query{PageSize: 1, PageNumber: 1})
	assert.Nil(t, err)
	assert.Equal(t, 1, total)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, name, apps[0].Name)

	apps, err = Mgr.GetByIDs(ctx, []uint{application.ID})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, name, apps[0].Name)

	assert.Equal(t, appGetByName.Name, apps[0].Name)
	err = Mgr.DeleteByID(ctx, appGetByName.ID)
	assert.Nil(t, err)
}

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.Application{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&membermodels.Member{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
