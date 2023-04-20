// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	groupdao "github.com/horizoncd/horizon/pkg/group/dao"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	membermanager "github.com/horizoncd/horizon/pkg/member"
	membermodels "github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/global"
	userdao "github.com/horizoncd/horizon/pkg/user/dao"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	callbacks "github.com/horizoncd/horizon/pkg/util/ormcallbacks"
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

func TestList(t *testing.T) {
	currentUser, err := common.UserFromContext(ctx)
	assert.Nil(t, err)
	userDAO := userdao.NewDAO(db)
	user2, err := userDAO.Create(ctx, &usermodels.User{
		Name:  "leo",
		Email: "leo@corp.com",
	})
	assert.Nil(t, err)
	assert.NotNil(t, user2)

	var (
		groupID         = 1
		name            = "application0"
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
	_, err = mgr.Create(ctx, application, map[string]string{user2.Email: role.Owner})
	assert.Nil(t, err)

	application = &models.Application{
		GroupID:         2,
		Name:            "application1",
		Description:     description,
		Priority:        priority,
		GitURL:          gitURL,
		GitSubfolder:    gitSubfolder,
		GitRef:          gitBranch,
		Template:        "springboot",
		TemplateRelease: "v1.2.3",
		CreatedBy:       createdBy,
		UpdatedBy:       updatedBy,
	}
	_, err = mgr.Create(ctx, application, nil)
	assert.Nil(t, err)
	total, _, err := mgr.List(ctx, []uint{1, 2}, q.New(nil))
	assert.Nil(t, err)
	assert.Equal(t, total, 2)

	k1 := make(map[string]interface{})
	k1[common.ApplicationQueryByTemplate] = template
	k1[common.ApplicationQueryByRelease] = templateRelease
	k1[common.ApplicationQueryName] = "app"
	total, _, err = mgr.List(ctx, []uint{1, 2}, q.New(k1))
	assert.Nil(t, err)
	assert.Equal(t, total, 1)

	k2 := make(map[string]interface{})
	k2[common.ApplicationQueryByUser] = user2.ID
	total, apps, err := mgr.List(ctx, nil, q.New(k2))
	assert.Nil(t, err)
	assert.Equal(t, total, 1)
	assert.Equal(t, apps[0].Name, "application0")
}
