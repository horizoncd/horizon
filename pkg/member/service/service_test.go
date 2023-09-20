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

package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	herror "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/orm"
	applicationmanagermock "github.com/horizoncd/horizon/mock/pkg/application/manager"
	clustermanagermock "github.com/horizoncd/horizon/mock/pkg/cluster/manager"
	groupmanagermock "github.com/horizoncd/horizon/mock/pkg/group/manager"
	pipelinemock "github.com/horizoncd/horizon/mock/pkg/pipelinerun/manager"
	rolemock "github.com/horizoncd/horizon/mock/pkg/rbac/role"
	applicationmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	memberctx "github.com/horizoncd/horizon/pkg/context"
	perror "github.com/horizoncd/horizon/pkg/errors"
	groupModels "github.com/horizoncd/horizon/pkg/group/models"
	"github.com/horizoncd/horizon/pkg/member/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	prmanager "github.com/horizoncd/horizon/pkg/pr/manager"
	pipelinemodels "github.com/horizoncd/horizon/pkg/pr/models"
	roleservice "github.com/horizoncd/horizon/pkg/rbac/role"
	"github.com/horizoncd/horizon/pkg/server/global"
	templatemodels "github.com/horizoncd/horizon/pkg/template/models"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	webhookmodels "github.com/horizoncd/horizon/pkg/webhook/models"
)

var (
	s       Service
	ctx     context.Context
	db      *gorm.DB
	manager *managerparam.Manager
)

func PostMemberEqualsMember(postMember PostMember, member *models.Member) bool {
	return models.ResourceType(postMember.ResourceType) == member.ResourceType &&
		postMember.ResourceID == member.ResourceID &&
		postMember.MemberInfo == member.MemberNameID &&
		postMember.MemberType == member.MemberType &&
		postMember.Role == member.Role
}

// nolint
func TestCreateAndUpdateGroupMember(t *testing.T) {
	createEnv(t)

	// mock the groupManager
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	groupManager := groupmanagermock.NewMockManager(mockCtrl)
	roleMockService := rolemock.NewMockService(mockCtrl)
	originService := &service{
		memberManager: manager.MemberMgr,
		groupManager:  groupManager,
		roleService:   roleMockService,
		userManager:   manager.UserMgr,
	}
	s = originService

	//  case  /group1/group2
	//    group1 member: tom(1), jerry(1), cat(1)
	//    group2 member: tom(2), jerry(2)
	var group1ID uint = 3
	var group2ID uint = 4
	var traversalIDs = "3,4"
	var tomID uint = 1
	var jerryID uint = 2
	var catID uint = 3

	users := []usermodels.User{
		{
			Model: global.Model{
				ID: tomID,
			},
			Name: "tom",
		},
		{
			Model: global.Model{
				ID: jerryID,
			},
			Name: "jerry",
		},
		{
			Model: global.Model{
				ID: catID,
			},
			Name: "cat",
		},
	}
	for i := range users {
		_, err := originService.userManager.Create(ctx, &users[i])
		assert.Nil(t, err)
	}
	defer func() {
		originService.userManager.DeleteUser(ctx, tomID)
		originService.userManager.DeleteUser(ctx, jerryID)
		originService.userManager.DeleteUser(ctx, catID)
	}()

	var grandUser userauth.User = &userauth.DefaultInfo{
		Name:     "tom",
		FullName: "tom",
		ID:       tomID,
	}
	ctx = context.WithValue(ctx, common.UserContextKey(), grandUser)

	// insert service to group2
	postMemberTom2 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group2ID,
		MemberInfo:   tomID,
		MemberType:   models.MemberUser,
		Role:         "owner",
	}
	member, err := originService.createMemberDirect(ctx, postMemberTom2)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberTom2, member))

	postMemberJerry2 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group2ID,
		MemberInfo:   jerryID,
		MemberType:   models.MemberUser,
		Role:         "owner",
	}
	member, err = originService.createMemberDirect(ctx, postMemberJerry2)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberJerry2, member))

	// insert member to group1
	postMemberTom1 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group1ID,
		MemberInfo:   tomID,
		MemberType:   models.MemberUser,
		Role:         "owner",
	}
	tomMember1, err := originService.createMemberDirect(ctx, postMemberTom1)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberTom1, tomMember1))

	postMemberJerry1 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group1ID,
		MemberInfo:   jerryID,
		MemberType:   models.MemberUser,
		Role:         "maintainer",
	}
	member, err = originService.createMemberDirect(ctx, postMemberJerry1)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberJerry1, member))

	postMemberCat1 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group1ID,
		MemberInfo:   catID,
		MemberType:   models.MemberUser,
		Role:         "maintainer",
	}
	member, err = originService.createMemberDirect(ctx, postMemberCat1)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberCat1, member))

	// test role smaller
	// create member exist  auto change to update role
	roleMockService.EXPECT().RoleCompare(ctx, gomock.Any(), gomock.Any()).Return(
		roleservice.RoleSmaller, nil).Times(1)
	// create member ok
	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).Times(2)
	groupManager.EXPECT().IsRootGroup(gomock.Any(), gomock.Any()).AnyTimes().Return(false)
	postMemberCat2 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group2ID,
		MemberInfo:   catID,
		MemberType:   models.MemberUser,
		Role:         "maintainer",
	}
	_, err = s.CreateMember(ctx, postMemberCat2)
	assert.Equal(t, perror.Cause(err), herror.ErrNoPrivilege)

	// create member exist  auto change to update role
	roleMockService.EXPECT().RoleCompare(ctx, gomock.Any(), gomock.Any()).Return(
		roleservice.RoleBigger, nil).AnyTimes()
	// create member ok
	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).Times(1)
	postMemberCat2 = PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group2ID,
		MemberInfo:   catID,
		MemberType:   models.MemberUser,
		Role:         "maintainer",
	}
	catMember2, err := s.CreateMember(ctx, postMemberCat2)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberCat2, catMember2))

	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).Times(2)
	postMemberCat2.Role = "develop"
	member, err = s.CreateMember(ctx, postMemberCat2)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberCat2, member))

	// update member not exist
	var memberIDNotExist uint = 123233434
	member, err = s.UpdateMember(ctx, memberIDNotExist, "owner")
	_, ok := perror.Cause(err).(*herror.HorizonErrNotFound)
	assert.True(t, ok)

	// update member correct
	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).Times(1)
	member, err = s.UpdateMember(ctx, tomMember1.ID, "maintainer")
	assert.Nil(t, err)
	assert.Equal(t, member.Role, "maintainer")
	assert.Equal(t, member.ID, tomMember1.ID)

	// remove member not exist
	err = s.RemoveMember(ctx, memberIDNotExist)
	_, ok = perror.Cause(err).(*herror.HorizonErrNotFound)
	assert.True(t, ok)

	// remove member ok
	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).Times(1)
	err = s.RemoveMember(ctx, catMember2.ID)
	assert.Nil(t, err)
}

// nolint
func TestListGroupMember(t *testing.T) {
	createEnv(t)

	// mock the groupManager
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	groupManager := groupmanagermock.NewMockManager(mockCtrl)
	originService := &service{
		memberManager: manager.MemberMgr,
		groupManager:  groupManager,
		userManager:   manager.UserMgr,
	}
	s = originService

	//  case  /group1/group2
	//    group1 service: tom(1), jerry(1), cat(1)
	//    group2 service: tom(2), jerry(2)
	//    ret: tom(2), jerry(2), cat(1)
	var group2ID uint = 2
	var group1ID uint = 1
	var traversalIDs = "1,2"
	var tomID uint = 1
	var jerryID uint = 2
	var catID uint = 3
	var grandUser userauth.User = &userauth.DefaultInfo{
		Name:     "tom",
		FullName: "tom",
		ID:       tomID,
	}
	users := []usermodels.User{
		{
			Model: global.Model{
				ID: tomID,
			},
			Name: "tom",
		},
		{
			Model: global.Model{
				ID: jerryID,
			},
			Name: "jerry",
		},
		{
			Model: global.Model{
				ID: catID,
			},
			Name: "cat",
		},
	}
	for i := range users {
		_, err := originService.userManager.Create(ctx, &users[i])
		assert.Nil(t, err)
	}
	defer func() {
		originService.userManager.DeleteUser(ctx, tomID)
		originService.userManager.DeleteUser(ctx, jerryID)
		originService.userManager.DeleteUser(ctx, catID)
	}()
	ctx = context.WithValue(ctx, common.UserContextKey(), grandUser)

	// insert service to group2
	postMemberTom2 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group2ID,
		MemberInfo:   tomID,
		MemberType:   models.MemberUser,
		Role:         "owner",
	}
	member, err := originService.createMemberDirect(ctx, postMemberTom2)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberTom2, member))

	postMemberJerry2 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group2ID,
		MemberInfo:   jerryID,
		MemberType:   models.MemberUser,
		Role:         "owner",
	}
	member, err = originService.createMemberDirect(ctx, postMemberJerry2)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberJerry2, member))

	// insert service to group1
	postMemberTom1 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group1ID,
		MemberInfo:   tomID,
		MemberType:   models.MemberUser,
		Role:         "owner",
	}
	member, err = originService.createMemberDirect(ctx, postMemberTom1)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberTom1, member))

	postMemberJerry1 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group1ID,
		MemberInfo:   jerryID,
		MemberType:   models.MemberUser,
		Role:         "maintainer",
	}
	member, err = originService.createMemberDirect(ctx, postMemberJerry1)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberJerry1, member))

	postMemberCat1 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group1ID,
		MemberInfo:   catID,
		MemberType:   models.MemberUser,
		Role:         "maintainer",
	}
	member, err = originService.createMemberDirect(ctx, postMemberCat1)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberCat1, member))

	// listmember of group2
	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).Times(1)
	groupManager.EXPECT().IsRootGroup(gomock.Any(), gomock.Any()).AnyTimes().Return(false)
	members, err := s.ListMember(ctx, common.ResourceGroup, group2ID)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(members))
	assert.True(t, PostMemberEqualsMember(postMemberTom2, &members[0]))
	assert.True(t, PostMemberEqualsMember(postMemberJerry2, &members[1]))
	assert.True(t, PostMemberEqualsMember(postMemberCat1, &members[2]))
}

func TestListApplicationMember(t *testing.T) {
	// TODO(tom)
}

func TestListApplicationInstanceMember(t *testing.T) {
	createEnv(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		group1ID       uint = 1
		group2ID       uint = 2
		application3ID uint = 3
		cluster4ID     uint = 4

		traversalIDs               = "1,2"
		sphID        uint          = 1
		jerryID      uint          = 2
		catID        uint          = 3
		catEmail                   = "cat@163.com"
		grandUser    userauth.User = &userauth.DefaultInfo{
			Name:     "sph",
			FullName: "sph",
			ID:       1,
		}
	)
	users := []usermodels.User{
		{
			Model: global.Model{
				ID: sphID,
			},
			Name: "sph",
		},
		{
			Model: global.Model{
				ID: jerryID,
			},
			Name: "jerry",
		},
		{
			Model: global.Model{
				ID: catID,
			},
			Email: catEmail,
			Name:  "cat",
		},
	}

	userManager := usermanager.New(db)
	for i := range users {
		_, err := userManager.Create(ctx, &users[i])
		assert.Nil(t, err)
	}
	defer func() {
		_ = userManager.DeleteUser(ctx, sphID)
		_ = userManager.DeleteUser(ctx, jerryID)
		_ = userManager.DeleteUser(ctx, catID)
	}()
	ctx = context.WithValue(ctx, common.UserContextKey(), grandUser) // nolint

	// mock the groupManager
	groupManager := groupmanagermock.NewMockManager(mockCtrl)
	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).Times(1)
	groupManager.EXPECT().IsRootGroup(gomock.Any(), gomock.Any()).AnyTimes().Return(false)

	// mock the applicationManager
	applicationManager := applicationmanagermock.NewMockManager(mockCtrl)
	applicationManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*applicationmodels.Application, error) {
		return &applicationmodels.Application{
			Model:       global.Model{},
			Name:        "",
			Description: "",
			GroupID:     group2ID,
		}, nil
	}).Times(1)

	// mock the applicationClusterManager
	clusterManager := clustermanagermock.NewMockManager(mockCtrl)
	clusterManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*clustermodels.Cluster, error) {
		return &clustermodels.Cluster{
			Model:         global.Model{},
			Name:          "",
			Description:   "",
			ApplicationID: application3ID,
		}, nil
	}).Times(1)

	originService := &service{
		memberManager:             manager.MemberMgr,
		groupManager:              groupManager,
		applicationManager:        applicationManager,
		applicationClusterManager: clusterManager,
	}
	s = originService

	// insert members
	postMembers := []PostMember{
		{
			ResourceType: common.ResourceGroup,
			ResourceID:   group1ID,
			MemberInfo:   sphID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceGroup,
			ResourceID:   group2ID,
			MemberInfo:   sphID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceGroup,
			ResourceID:   group2ID,
			MemberInfo:   jerryID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceApplication,
			ResourceID:   application3ID,
			MemberInfo:   sphID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceApplication,
			ResourceID:   application3ID,
			MemberInfo:   catID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceCluster,
			ResourceID:   cluster4ID,
			MemberInfo:   catID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
	}

	for _, postMember := range postMembers {
		result, err := originService.createMemberDirect(ctx, postMember)
		assert.Nil(t, err)
		assert.True(t, PostMemberEqualsMember(postMember, result))
	}

	// check members
	members, err := s.ListMember(ctx, common.ResourceCluster, cluster4ID)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(members))
	assert.True(t, PostMemberEqualsMember(postMembers[5], &members[0]))
	assert.True(t, PostMemberEqualsMember(postMembers[3], &members[1]))
	assert.True(t, PostMemberEqualsMember(postMembers[2], &members[2]))

	ctx = context.WithValue(ctx, memberctx.MemberQueryOnCondition, true)
	ctx = context.WithValue(ctx, memberctx.MemberDirectMemberOnly, true)
	ctx = context.WithValue(ctx, memberctx.MemberEmails, []string{catEmail})
	members, err = s.ListMember(ctx, common.ResourceCluster, cluster4ID)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMembers[5], &members[0]))
}

//	 case  /group1/group2/application/cluster
//			group1 member: sph(1)
//			group2 member: sph(2), jerry(2)
//			application3 member: sph(3), cat(3)
//			cluster4 member: cat(4)
//			ret: sph(3), jerry(2), cat(4)
//
// nolint
func TestGetPipelinerunAndCheckrunMember(t *testing.T) {
	createEnv(t)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	var (
		group1ID       uint = 1
		group2ID       uint = 2
		application3ID uint = 3
		cluster4ID     uint = 4

		traversalIDs               = "1,2"
		sphID        uint          = 1
		jerryID      uint          = 2
		catID        uint          = 3
		grandUser    userauth.User = &userauth.DefaultInfo{
			Name:     "sph",
			FullName: "sph",
			ID:       1,
		}
		pipelineRunID uint = 23123
		checkrunID    uint = 12138
	)
	users := []usermodels.User{
		{
			Model: global.Model{
				ID: sphID,
			},
			Name: "sph",
		},
		{
			Model: global.Model{
				ID: jerryID,
			},
			Name: "jerry",
		},
		{
			Model: global.Model{
				ID: catID,
			},
			Name: "cat",
		},
	}

	userManager := usermanager.New(db)
	for i := range users {
		_, err := userManager.Create(ctx, &users[i])
		assert.Nil(t, err)
	}
	defer func() {
		userManager.DeleteUser(ctx, sphID)
		userManager.DeleteUser(ctx, jerryID)
		userManager.DeleteUser(ctx, catID)
	}()
	ctx = context.WithValue(ctx, common.UserContextKey(), grandUser)

	// mock the groupManager
	groupManager := groupmanagermock.NewMockManager(mockCtrl)
	groupManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*groupModels.Group, error) {
		return &groupModels.Group{
			Model:           global.Model{},
			Name:            "",
			Path:            "",
			VisibilityLevel: "",
			Description:     "",
			ParentID:        0,
			TraversalIDs:    traversalIDs,
		}, nil
	}).AnyTimes()
	groupManager.EXPECT().IsRootGroup(gomock.Any(), gomock.Any()).AnyTimes().Return(false)

	// mock the applicationManager
	applicationManager := applicationmanagermock.NewMockManager(mockCtrl)
	applicationManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*applicationmodels.Application, error) {
		return &applicationmodels.Application{
			Model:       global.Model{},
			Name:        "",
			Description: "",
			GroupID:     group2ID,
		}, nil
	}).AnyTimes()

	// mock the applicationClusterManager
	clusterManager := clustermanagermock.NewMockManager(mockCtrl)
	clusterManager.EXPECT().GetByID(gomock.Any(),
		gomock.Any()).DoAndReturn(func(_ context.Context, id uint) (*clustermodels.Cluster, error) {
		return &clustermodels.Cluster{
			Model:         global.Model{},
			Name:          "",
			Description:   "",
			ApplicationID: application3ID,
		}, nil
	}).AnyTimes()

	pipelineMockManager := pipelinemock.NewMockPipelineRunManager(mockCtrl)
	pipelineMockManager.EXPECT().GetByID(gomock.Any(), pipelineRunID).Return(&pipelinemodels.Pipelinerun{
		ID:        0,
		ClusterID: cluster4ID,
	}, nil).AnyTimes()

	checkMockManager := pipelinemock.NewMockCheckManager(mockCtrl)
	checkMockManager.EXPECT().GetCheckRunByID(gomock.Any(), checkrunID).
		Return(&pipelinemodels.CheckRun{
			PipelineRunID: pipelineRunID,
		}, nil).AnyTimes()

	roleSvc := rolemock.NewMockService(mockCtrl)
	originService := &service{
		memberManager:             manager.MemberMgr,
		groupManager:              groupManager,
		applicationManager:        applicationManager,
		applicationClusterManager: clusterManager,
		prMgr: &prmanager.PRManager{
			PipelineRun: pipelineMockManager,
			Check:       checkMockManager,
		},
		roleService: roleSvc,
		userManager: manager.UserMgr,
	}
	s = originService

	roleSvc.EXPECT().RoleCompare(gomock.Any(), roleservice.Owner, roleservice.Maintainer).Return(roleservice.RoleBigger, nil)
	roleSvc.EXPECT().RoleCompare(gomock.Any(), roleservice.Owner, roleservice.Owner).Return(roleservice.RoleEqual, nil)
	roleSvc.EXPECT().RoleCompare(gomock.Any(), roleservice.Maintainer, roleservice.Maintainer).Return(roleservice.RoleEqual, nil)

	// insert members
	postMembers := []PostMember{
		{
			ResourceType: common.ResourceGroup,
			ResourceID:   group1ID,
			MemberInfo:   sphID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceGroup,
			ResourceID:   group2ID,
			MemberInfo:   sphID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceGroup,
			ResourceID:   group2ID,
			MemberInfo:   jerryID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceApplication,
			ResourceID:   application3ID,
			MemberInfo:   sphID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceApplication,
			ResourceID:   application3ID,
			MemberInfo:   catID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
		{
			ResourceType: common.ResourceCluster,
			ResourceID:   cluster4ID,
			MemberInfo:   catID,
			MemberType:   models.MemberUser,
			Role:         "owner",
		},
	}

	for _, postMember := range postMembers {
		result, err := originService.createMemberDirect(ctx, postMember)
		assert.Nil(t, err)
		assert.True(t, PostMemberEqualsMember(postMember, result))
	}

	// check members
	pipelineRunIDStr := strconv.FormatUint(uint64(pipelineRunID), 10)
	members, err := s.GetMemberOfResource(ctx, common.ResourcePipelinerun, pipelineRunIDStr)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMembers[3], members))

	members, err = s.GetMemberOfResource(ctx, common.ResourceCheckrun,
		strconv.FormatUint(uint64(checkrunID), 10))
	assert.NoError(t, err)
	assert.True(t, PostMemberEqualsMember(postMembers[3], members))

	members, err = s.UpdateMember(ctx, members.ID, roleservice.Maintainer)
	assert.Nil(t, err)
	assert.Equal(t, roleservice.Maintainer, members.Role)

	err = s.RemoveMember(ctx, members.ID)
	assert.Nil(t, err)
}

func TestListTemplateMember(t *testing.T) {
	createEnv(t)

	s = &service{
		memberManager:             manager.MemberMgr,
		groupManager:              manager.GroupMgr,
		applicationManager:        manager.ApplicationMgr,
		applicationClusterManager: manager.ClusterMgr,
		templateManager:           manager.TemplateMgr,
		prMgr:                     manager.PRMgr,
		roleService:               nil,
	}

	jerry := &usermodels.User{
		Name:     "Jerry",
		FullName: "Jerry",
		Email:    "jerry@mail.com",
		Phone:    "",
		Admin:    false,
	}
	jerry, err := manager.UserMgr.Create(ctx, jerry)
	assert.Nil(t, err)

	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:     jerry.Name,
		FullName: jerry.FullName,
		ID:       jerry.ID,
		Email:    jerry.Email,
		Admin:    jerry.Admin,
	})

	group1 := &groupModels.Group{
		Name:            "test",
		Path:            "test",
		VisibilityLevel: "",
		Description:     "test",
		ParentID:        0,
		TraversalIDs:    "1",
	}
	group1, err = manager.GroupMgr.Create(ctx, group1)
	assert.Nil(t, err)

	group2 := &groupModels.Group{
		Name:            "test2",
		Path:            "test2",
		VisibilityLevel: "",
		Description:     "test2",
		ParentID:        group1.ID,
		TraversalIDs:    "1,2",
	}
	group2, err = manager.GroupMgr.Create(ctx, group2)
	assert.Nil(t, err)

	template := &templatemodels.Template{
		Name:       "javaapp",
		Repository: "repo",
		GroupID:    group2.ID,
	}
	template, err = manager.TemplateMgr.Create(ctx, template)
	assert.Nil(t, err)

	_, err = manager.MemberMgr.Create(ctx, &models.Member{
		ResourceType: models.TypeGroup,
		ResourceID:   group1.ID,
		Role:         "pe",
		MemberType:   0,
		MemberNameID: jerry.ID,
	})
	assert.Nil(t, err)
	err = manager.MemberMgr.DeleteMember(ctx, 1)
	assert.Nil(t, err)
	err = manager.MemberMgr.DeleteMember(ctx, 2)
	assert.Nil(t, err)

	memberInfo, err := s.GetMemberOfResource(ctx, common.ResourceTemplate, fmt.Sprintf("%d", template.ID))
	assert.Nil(t, err)
	assert.NotNil(t, memberInfo)
	assert.Equal(t, jerry.ID, memberInfo.MemberNameID)
	assert.Equal(t, "pe", memberInfo.Role)
}

func TestListWebhookMember(t *testing.T) {
	createEnv(t)

	s = &service{
		memberManager:             manager.MemberMgr,
		groupManager:              manager.GroupMgr,
		applicationManager:        manager.ApplicationMgr,
		applicationClusterManager: manager.ClusterMgr,
		templateManager:           manager.TemplateMgr,
		prMgr:                     manager.PRMgr,
		roleService:               nil,
		webhookManager:            manager.WebhookMgr,
	}

	jerry := &usermodels.User{
		Name:     "Jerry",
		FullName: "Jerry",
		Email:    "jerry@mail.com",
		Phone:    "",
		Admin:    false,
	}
	jerry, err := manager.UserMgr.Create(ctx, jerry)
	assert.Nil(t, err)

	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:     jerry.Name,
		FullName: jerry.FullName,
		ID:       jerry.ID,
		Email:    jerry.Email,
		Admin:    jerry.Admin,
	})

	group1 := &groupModels.Group{
		Name:            "test",
		Path:            "test",
		VisibilityLevel: "",
		Description:     "test",
		ParentID:        0,
		TraversalIDs:    "1",
	}
	group1, err = manager.GroupMgr.Create(ctx, group1)
	assert.Nil(t, err)

	webhook := &webhookmodels.Webhook{
		URL:          "test",
		ResourceType: "groups",
		ResourceID:   group1.ID,
	}
	webhook, err = manager.WebhookMgr.CreateWebhook(ctx, webhook)
	assert.Nil(t, err)

	webhookLog := &webhookmodels.WebhookLog{
		WebhookID: webhook.ID,
	}
	webhookLog, err = manager.WebhookMgr.CreateWebhookLog(ctx, webhookLog)
	assert.Nil(t, err)

	_, err = manager.MemberMgr.Create(ctx, &models.Member{
		ResourceType: models.TypeGroup,
		ResourceID:   group1.ID,
		Role:         "pe",
		MemberType:   0,
		MemberNameID: jerry.ID,
	})
	assert.Nil(t, err)

	err = manager.MemberMgr.DeleteMember(ctx, 1)
	assert.Nil(t, err)

	memberInfo, err := s.GetMemberOfResource(ctx, common.ResourceWebhook, fmt.Sprintf("%d", webhook.ID))
	assert.Nil(t, err)
	assert.NotNil(t, memberInfo)
	assert.Equal(t, jerry.ID, memberInfo.MemberNameID)
	assert.Equal(t, "pe", memberInfo.Role)

	memberInfo, err = s.GetMemberOfResource(ctx, common.ResourceWebhookLog, fmt.Sprintf("%d", webhookLog.ID))
	assert.Nil(t, err)
	assert.NotNil(t, memberInfo)
	assert.Equal(t, jerry.ID, memberInfo.MemberNameID)
	assert.Equal(t, "pe", memberInfo.Role)
}

func createEnv(t *testing.T) {
	db, _ = orm.NewSqliteDB("")
	err := db.AutoMigrate(&models.Member{},
		&groupModels.Group{},
		&usermodels.User{},
		&applicationmodels.Application{},
		&clustermodels.Cluster{},
		&templatemodels.Template{},
		&webhookmodels.Webhook{},
		&webhookmodels.WebhookLog{},
	)

	assert.Nil(t, err)

	ctx = context.Background()
	manager = managerparam.InitManager(db)
}
