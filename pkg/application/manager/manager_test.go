package manager

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"g.hz.netease.com/horizon/pkg/util/errors"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func Test(t *testing.T) {
	userDAO := userdao.NewDAO()
	user1, err := userDAO.Create(ctx, &usermodels.User{
		Name:  "tony",
		Email: "tony@corp.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user1)

	user2, err := userDAO.Create(ctx, &usermodels.User{
		Name:  "leo",
		Email: "leo@corp.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user2)

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
		createdBy       = user1.ID
		updatedBy       = user1.ID
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
	application, err = Mgr.Create(ctx, application, []string{user2.Email})
	assert.Nil(t, err)

	assert.Equal(t, name, application.Name)
	assert.Equal(t, 1, int(application.ID))
	clusterMembers, err := member.Mgr.ListDirectMember(ctx, membermodels.TypeApplication, application.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(clusterMembers))
	assert.Equal(t, user2.ID, clusterMembers[1].MemberNameID)
	assert.Equal(t, role.Owner, clusterMembers[1].Role)

	application2 := &models.Application{
		Name: "application2",
	}
	application2, err2 := Mgr.Create(ctx, application2, []string{user2.Email, "not-exist@corp.com"})
	assert.Nil(t, application2)
	assert.NotNil(t, err2)
	t.Logf("%v", err2)
	assert.Equal(t, http.StatusNotFound, errors.Status(err2))

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
	if err := db.AutoMigrate(&membermodels.Member{}, &usermodels.User{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
