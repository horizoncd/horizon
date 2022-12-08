package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	groupdao "g.hz.netease.com/horizon/pkg/group/dao"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	membermanager "g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/global"
	userdao "g.hz.netease.com/horizon/pkg/user/dao"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	callbacks "g.hz.netease.com/horizon/pkg/util/ormcallbacks"
	"github.com/stretchr/testify/assert"
)

var (
	db, _     = orm.NewSqliteDB("")
	ctx       context.Context
	mgr       = New(db)
	memberMgr = membermanager.New(db)
)

func TestMain(m *testing.M) {
	currentUser := usermodels.User{
		Model: global.Model{
			ID: 110,
		},
		Name:  "tony",
		Email: "tony@corp.com",
	}
	// nolint
	db = db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name: currentUser.Name,
		ID:   currentUser.ID,
	}))
	callbacks.RegisterCustomCallbacks(db)
	// nolint
	ctx = context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name: currentUser.Name,
		ID:   currentUser.ID,
	})

	if err := db.AutoMigrate(&models.Application{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&membermodels.Member{}, &usermodels.User{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&groupmodels.Group{}); err != nil {
		panic(err)
	}

	userDAO := userdao.NewDAO(db)
	_, err := userDAO.Create(ctx, &currentUser)
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func Test(t *testing.T) {
	currentUser, err := common.UserFromContext(ctx)
	assert.Nil(t, err)

	userDAO := userdao.NewDAO(db)
	user2, err := userDAO.Create(ctx, &usermodels.User{
		Name:  "leo",
		Email: "leo@corp.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user2)

	groupDAO := groupdao.NewDAO(db)
	group, err := groupDAO.Create(ctx, &groupmodels.Group{
		Model:    global.Model{},
		Name:     "group1",
		Path:     "/group1",
		ParentID: 0,
	})
	assert.Nil(t, err)
	assert.NotNil(t, group)

	group1, err := groupDAO.Create(ctx, &groupmodels.Group{
		Model:    global.Model{},
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
		createdBy       = currentUser.GetID()
		updatedBy       = currentUser.GetID()
	)
	application := &models.Application{
		GroupID:         uint(groupID),
		Name:            name,
		Description:     description,
		Priority:        priority,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitBranch,
		Template:        template,
		TemplateRelease: templateRelease,
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
	}
	application, err = mgr.Create(ctx, application, map[string]string{user2.Email: role.Owner})
	assert.Nil(t, err)

	assert.Equal(t, name, application.Name)
	assert.Equal(t, 1, int(application.ID))
	clusterMembers, err := memberMgr.ListDirectMember(ctx, membermodels.TypeApplication, application.ID)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(clusterMembers))
	assert.Equal(t, user2.ID, clusterMembers[1].MemberNameID)
	assert.Equal(t, role.Owner, clusterMembers[1].Role)

	b, err := json.Marshal(application)
	assert.Nil(t, err)
	t.Logf(string(b))

	appGetByID, err := mgr.GetByID(ctx, application.ID)
	assert.Nil(t, err)
	assert.Equal(t, application.Name, appGetByID.Name)

	appGetByName, err := mgr.GetByName(ctx, application.Name)
	assert.Nil(t, err)
	assert.Equal(t, application.ID, appGetByName.ID)

	appGetByName.Description = "new"
	appGetByName, err = mgr.UpdateByID(ctx, application.ID, appGetByName)
	assert.Nil(t, err)

	apps, err := mgr.GetByNameFuzzily(ctx, "app")
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, appGetByName.Name, apps[0].Name)

	apps, err = mgr.GetByIDs(ctx, []uint{application.ID})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, name, apps[0].Name)

	apps, err = mgr.GetByGroupIDs(ctx, []uint{uint(groupID)})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, name, apps[0].Name)

	// test transfer application
	var transferGroupID uint = 2
	err = mgr.Transfer(ctx, application.ID, transferGroupID)
	assert.Nil(t, err)

	var transferNotExistGroupID uint = 100
	err = mgr.Transfer(ctx, application.ID, transferNotExistGroupID)
	assert.NotNil(t, err)

	// case 2 create the group and retry ok
	apps, err = mgr.GetByIDs(ctx, []uint{application.ID})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(apps))
	assert.Equal(t, transferGroupID, apps[0].GroupID)

	assert.Equal(t, appGetByName.Name, apps[0].Name)
	err = mgr.DeleteByID(ctx, appGetByName.ID)
	assert.Nil(t, err)

	app, err := mgr.GetByIDIncludeSoftDelete(ctx, appGetByName.ID)
	assert.Nil(t, err)
	assert.NotNil(t, app)
}
