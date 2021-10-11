package application

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	orm2 "g.hz.netease.com/horizon/pkg/lib/orm"
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
		priority        = P0
		gitURL          = "ssh://git@github.com"
		gitSubfolder    = "/"
		gitBranch       = "develop"
		template        = "javaapp"
		templateRelease = "v1.1.0"
		createdBy       = "tony"
		updatedBy       = "tony"
	)
	application := &Application{
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
}

func TestMain(m *testing.M) {
	db, _ = orm2.NewSqliteDB("")
	if err := db.AutoMigrate(&Application{}); err != nil {
		panic(err)
	}
	ctx = orm2.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
