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

package member

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/group"
	"github.com/horizoncd/horizon/lib/orm"
	rolemock "github.com/horizoncd/horizon/mock/pkg/rbac/role"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	appmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	roleservice "github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/global"
	service "github.com/horizoncd/horizon/pkg/service"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	// use tmp sqlite
	db, _                    = orm.NewSqliteDB("")
	ctx                      = context.TODO()
	contextUserID       uint = 1
	contextUserName          = "Tony"
	contextUserFullName      = "TonyWu"
	manager                  = managerparam.InitManager(db)
	groupCtl            group.Controller
	groupSvc            service.GroupService
	applicationSvc      service.ApplicationService
	clusterSvc          service.ClusterService
)

var (
	user1ID   uint = 1
	user1Name      = contextUserName

	user2ID   uint = 2
	user2Name      = "tom"

	user3Name = "jerry"

	user4Name = "alias"

	user5Name = "henry"
)

// nolint
func createContext(t *testing.T) {
	db, _ = orm.NewSqliteDB("")
	manager = managerparam.InitManager(db)

	ctx = context.WithValue(context.Background(), common.UserContextKey(), &userauth.DefaultInfo{
		Name:     contextUserName,
		FullName: contextUserFullName,
		ID:       contextUserID,
		Admin:    true,
	})
	// create table
	err := db.AutoMigrate(&appmodels.Group{}, &appmodels.User{},
		&appmodels.Application{}, &appmodels.Member{},
		&appmodels.Template{}, &appmodels.TemplateRelease{})
	assert.Nil(t, err)

	groupCtl = group.NewController(&param.Param{Manager: manager})
	groupSvc = service.NewGroupService(manager)
	applicationSvc = service.NewApplicationService(groupSvc, manager)
	clusterSvc = service.NewClusterService(applicationSvc, manager)
}

func MemberSame(m1, m2 Member) bool {
	if m1.MemberName == m2.MemberName &&
		m1.MemberNameID == m2.MemberNameID &&
		m1.ResourceType == m2.ResourceType &&
		m1.ResourceID == m2.ResourceID &&
		m1.Role == m2.Role &&
		m1.GrantedBy == m2.GrantedBy {
		return true
	}
	return false
}

func CreateUsers(t *testing.T) {
	// create user
	user1 := appmodels.User{
		Model: global.Model{},
		Name:  user1Name,
	}

	user2 := user1
	user2.Name = user2Name

	user3 := user1
	user3.Name = user3Name

	user4 := user1
	user4.Name = user4Name

	user5 := user1
	user5.Name = user5Name

	for _, user := range []appmodels.User{user1, user2, user3, user4, user5} {
		_, err := manager.UserManager.Create(ctx, &user)
		assert.Nil(t, err)
	}
}

func TestCreateGroupWithOwner(t *testing.T) {
	createContext(t)
	memberService := service.NewMemberService(nil, nil, manager)
	ctl := NewController(&param.Param{
		MemberService:  memberService,
		Manager:        manager,
		GroupSvc:       groupSvc,
		ApplicationSvc: applicationSvc,
		ClusterSvc:     clusterSvc,
	})

	CreateUsers(t)

	// create group
	newGroup := &group.NewGroup{
		Name:            "1",
		Path:            "1",
		VisibilityLevel: "private",
		Description:     "i am a private Group",
		ParentID:        0,
	}

	groupID, err := groupCtl.CreateGroup(ctx, newGroup)
	assert.Nil(t, err)

	retMembers, err := ctl.ListMember(ctx, common.ResourceGroup, groupID)
	expectMember := Member{
		MemberType:   appmodels.MemberUser,
		MemberName:   contextUserName,
		MemberNameID: contextUserID,
		ResourceType: appmodels.TypeGroup,
		ResourceID:   groupID,
		Role:         roleservice.Owner,
		GrantedBy:    contextUserID,
	}
	assert.Nil(t, err)
	assert.NotNil(t, retMembers)
	assert.True(t, MemberSame(retMembers[0], expectMember))

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&appmodels.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&appmodels.Member{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&appmodels.User{})
}

func PostMemberAndMemberEqual(postMember PostMember, member2 Member) bool {
	return postMember.ResourceType == string(member2.ResourceType) &&
		postMember.ResourceID == member2.ResourceID &&
		postMember.MemberType == member2.MemberType &&
		postMember.MemberNameID == member2.MemberNameID &&
		postMember.Role == member2.Role
}

func TestCreateGetUpdateRemoveList(t *testing.T) {
	createContext(t)
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	roleMockService := rolemock.NewMockService(mockCtrl)
	CreateUsers(t)

	// mock the RoleCompare
	roleMockService.EXPECT().RoleCompare(ctx, gomock.Any(), gomock.Any()).Return(
		roleservice.RoleBigger, nil).AnyTimes()

	memberService := service.NewMemberService(roleMockService, nil, manager)
	ctl := NewController(&param.Param{
		MemberService:  memberService,
		Manager:        manager,
		GroupSvc:       groupSvc,
		ApplicationSvc: applicationSvc,
		ClusterSvc:     clusterSvc,
	})

	// create group
	newGroup := &group.NewGroup{
		Name:            "1",
		Path:            "1",
		VisibilityLevel: "private",
		Description:     "i am a private Group",
		ParentID:        0,
	}

	groupID, err := groupCtl.CreateGroup(ctx, newGroup)
	assert.Nil(t, err)

	// create member
	postMember2 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   groupID,
		MemberNameID: user2ID,
		MemberType:   appmodels.MemberUser,
		Role:         roleservice.Owner,
	}
	retMember2, err := ctl.CreateMember(ctx, &postMember2)
	assert.Nil(t, err)
	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember2))

	// update member
	retMember3, err := ctl.UpdateMember(ctx, retMember2.ID, "maitainer")
	assert.Nil(t, err)
	postMember2.Role = "maitainer"

	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember3))

	// remove the member
	err = ctl.RemoveMember(ctx, retMember2.ID)
	assert.Nil(t, err)

	// list member (create post2 and then list)
	retMember2, err = ctl.CreateMember(ctx, &postMember2)
	assert.Nil(t, err)
	assert.True(t, PostMemberAndMemberEqual(postMember2, *retMember2))

	postMemberOwner := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   groupID,
		MemberNameID: user1ID,
		MemberType:   appmodels.MemberUser,
		Role:         roleservice.Owner,
	}
	members, err := ctl.ListMember(ctx, common.ResourceGroup, groupID)
	assert.Nil(t, err)
	assert.Equal(t, len(members), 2)
	assert.True(t, PostMemberAndMemberEqual(postMemberOwner, members[0]))
	assert.True(t, PostMemberAndMemberEqual(postMember2, members[1]))

	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&appmodels.Group{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&appmodels.Member{})
	db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&appmodels.User{})
}

func TestTemplateMember(t *testing.T) {
	createContext(t)
	memberService := service.NewMemberService(nil, nil, manager)
	ctl := NewController(&param.Param{
		MemberService:  memberService,
		Manager:        manager,
		GroupSvc:       groupSvc,
		ApplicationSvc: applicationSvc,
		ClusterSvc:     clusterSvc,
	})

	onlyOwner := false
	template := &appmodels.Template{
		Model:     global.Model{},
		Name:      "javaapp",
		ChartName: "javaapp",
		GroupID:   0,
		OnlyOwner: &onlyOwner,
	}
	template, err := manager.TemplateMgr.Create(ctx, template)
	assert.Nil(t, err)

	recommended := true
	release := &appmodels.TemplateRelease{
		Model:        global.Model{},
		Template:     1,
		TemplateName: "javaapp",
		ChartVersion: "v1.0.0",
		ChartName:    "javaapp",
		Recommended:  &recommended,
		OnlyOwner:    &onlyOwner,
	}

	_, err = manager.TemplateReleaseManager.Create(ctx, release)
	assert.Nil(t, err)

	jerry := &appmodels.User{
		Model:    global.Model{},
		Name:     "Jerry",
		FullName: "Jerry",
		Email:    "Jerry@mail.com",
		Phone:    "12390821",
		Admin:    false,
	}
	jerry, err = manager.UserManager.Create(ctx, jerry)
	assert.Nil(t, err)
	createMemberParam := &PostMember{
		ResourceType: common.ResourceTemplate,
		ResourceID:   template.ID,
		MemberType:   appmodels.MemberUser,
		MemberNameID: jerry.ID,
		Role:         "owner",
	}
	_, err = ctl.CreateMember(ctx, createMemberParam)
	assert.Nil(t, err)

	member, err := ctl.GetMemberOfResource(ctx, common.ResourceTemplate, template.ID)
	assert.Nil(t, err)
	assert.Equal(t, createMemberParam.Role, member.Role)
}
