package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"

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

	groupDAO := groupdao.NewDAO()
	group, err := groupDAO.Create(ctx, &groupmodels.Group{
		Model:    gorm.Model{},
		Name:     "group1",
		Path:     "/group1",
		ParentID: 0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, group)

	group1, err := groupDAO.Create(ctx, &groupmodels.Group{
		Model:    gorm.Model{},
		Name:     "group2",
		Path:     "/group2",
		ParentID: 0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, group1)

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
	application, err = Mgr.Create(ctx, application, map[string]string{user2.Email: role.Owner})
	assert.Nil(t, err)

	assert.Equal(t, name, application.Name)
	assert.Equal(t, 1, int(application.ID))
	clusterMembers, err := member.Mgr.ListDirectMember(ctx, membermodels.TypeApplication, application.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(clusterMembers))
	assert.Equal(t, user2.ID, clusterMembers[1].MemberNameID)
	assert.Equal(t, role.Owner, clusterMembers[1].Role)

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

	totalForUser, appsForUser, err := Mgr.ListUserAuthorizedByNameFuzzily(ctx,
		"app", []uint{1}, user2.ID, &q.Query{
			PageNumber: 0,
			PageSize:   common.DefaultPageSize,
		})
	assert.Nil(t, err)
	assert.Equal(t, 1, totalForUser)
	assert.Equal(t, 1, len(appsForUser))
	assert.Equal(t, name, apps[0].Name)

	apps, err = Mgr.GetByIDs(ctx, []uint{application.ID})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, name, apps[0].Name)

	apps, err = Mgr.GetByGroupIDs(ctx, []uint{uint(groupID)})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, name, apps[0].Name)

	// test transfer application
	var transferGroupID uint = 2
	err = Mgr.Transfer(ctx, application.ID, transferGroupID)
	assert.Nil(t, err)

	var transferNotExistGroupID uint = 100
	err = Mgr.Transfer(ctx, application.ID, transferNotExistGroupID)
	assert.NotNil(t, err)

	// case 2 create the group and retry ok
	apps, err = Mgr.GetByIDs(ctx, []uint{application.ID})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, transferGroupID, apps[0].GroupID)

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
	if err := db.AutoMigrate(&groupmodels.Group{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}
