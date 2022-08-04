package service

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	applicationmanagermock "g.hz.netease.com/horizon/mock/pkg/application/manager"
	clustermanagermock "g.hz.netease.com/horizon/mock/pkg/cluster/manager"
	groupmanagermock "g.hz.netease.com/horizon/mock/pkg/group/manager"
	pipelinemock "g.hz.netease.com/horizon/mock/pkg/pipelinerun/manager"
	rolemock "g.hz.netease.com/horizon/mock/pkg/rbac/role"
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	memberctx "g.hz.netease.com/horizon/pkg/context"
	groupModels "g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	pipelinemodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	roleservice "g.hz.netease.com/horizon/pkg/rbac/role"
	"g.hz.netease.com/horizon/pkg/server/global"
	templatemodels "g.hz.netease.com/horizon/pkg/template/models"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
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
		memberManager: manager.MemberManager,
		groupManager:  groupManager,
		roleService:   roleMockService,
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
	}).Times(1)
	groupManager.EXPECT().IsRootGroup(gomock.Any(), gomock.Any()).AnyTimes().Return(false)
	postMemberCat2 := PostMember{
		ResourceType: common.ResourceGroup,
		ResourceID:   group2ID,
		MemberInfo:   catID,
		MemberType:   models.MemberUser,
		Role:         "maintainer",
	}
	_, err = s.CreateMember(ctx, postMemberCat2)
	assert.Equal(t, err, ErrGrantHighRole)

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
	}).Times(1)
	postMemberCat2.Role = "develop"
	member, err = s.CreateMember(ctx, postMemberCat2)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMemberCat2, member))

	// update member not exist
	var memberIDNotExist uint = 123233434
	member, err = s.UpdateMember(ctx, memberIDNotExist, "owner")
	assert.Equal(t, err.Error(), ErrMemberNotExist.Error())

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
	assert.Equal(t, err.Error(), ErrMemberNotExist.Error())

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
		memberManager: manager.MemberManager,
		groupManager:  groupManager,
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
		memberManager:             manager.MemberManager,
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

	userMgr := usermanager.New(db)
	_, err = userMgr.Create(ctx, &usermodels.User{Model: global.Model{ID: catID}, Email: catEmail})
	assert.Nil(t, err)

	ctx = context.WithValue(ctx, memberctx.MemberQueryOnCondition, true)
	ctx = context.WithValue(ctx, memberctx.MemberDirectMemberOnly, true)
	ctx = context.WithValue(ctx, memberctx.MemberEmails, []string{catEmail})
	members, err = s.ListMember(ctx, common.ResourceCluster, cluster4ID)
	assert.Nil(t, err)
	assert.True(t, PostMemberEqualsMember(postMembers[5], &members[0]))
}

//  case  /group1/group2/application/cluster
//		group1 member: sph(1)
//		group2 member: sph(2), jerry(2)
//		application3 member: sph(3), cat(3)
//		cluster4 member: cat(4)
//		ret: sph(3), jerry(2), cat(4)
// nolint
func TestGetPipelinerunMember(t *testing.T) {
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
	)
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

	pipelineMockManager := pipelinemock.NewMockManager(mockCtrl)
	pipelineMockManager.EXPECT().GetByID(gomock.Any(), pipelineRunID).Return(&pipelinemodels.Pipelinerun{
		ID:        0,
		ClusterID: cluster4ID,
	}, nil).AnyTimes()

	roleSvc := rolemock.NewMockService(mockCtrl)
	originService := &service{
		memberManager:             manager.MemberManager,
		groupManager:              groupManager,
		applicationManager:        applicationManager,
		applicationClusterManager: clusterManager,
		pipelineManager:           pipelineMockManager,
		roleService:               roleSvc,
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

	members, err = s.UpdateMember(ctx, members.ID, roleservice.Maintainer)
	assert.Nil(t, err)
	assert.Equal(t, roleservice.Maintainer, members.Role)

	err = s.RemoveMember(ctx, members.ID)
	assert.Nil(t, err)
}

func TestListTemplateMember(t *testing.T) {
	createEnv(t)

	s = &service{
		memberManager:             manager.MemberManager,
		groupManager:              manager.GroupManager,
		applicationManager:        manager.ApplicationManager,
		applicationClusterManager: manager.ClusterMgr,
		templateManager:           manager.TemplateMgr,
		pipelineManager:           manager.PipelinerunMgr,
		roleService:               nil,
	}

	jerry := &usermodels.User{
		Name:     "Jerry",
		FullName: "Jerry",
		Email:    "jerry@mail.com",
		Phone:    "",
		OIDCId:   "HZjerry",
		OIDCType: "netease",
		Admin:    false,
	}
	jerry, err := manager.UserManager.Create(ctx, jerry)
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
	group1, err = manager.GroupManager.Create(ctx, group1)
	assert.Nil(t, err)

	group2 := &groupModels.Group{
		Name:            "test2",
		Path:            "test2",
		VisibilityLevel: "",
		Description:     "test2",
		ParentID:        group1.ID,
		TraversalIDs:    "1,2",
	}
	group2, err = manager.GroupManager.Create(ctx, group2)
	assert.Nil(t, err)

	template := &templatemodels.Template{
		Name:       "javaapp",
		Repository: "repo",
		GroupID:    group2.ID,
	}
	template, err = manager.TemplateMgr.Create(ctx, template)
	assert.Nil(t, err)

	_, err = manager.MemberManager.Create(ctx, &models.Member{
		ResourceType: models.TypeGroup,
		ResourceID:   group1.ID,
		Role:         "pe",
		MemberType:   0,
		MemberNameID: jerry.ID,
	})
	assert.Nil(t, err)
	err = manager.MemberManager.DeleteMember(ctx, 1)
	assert.Nil(t, err)
	err = manager.MemberManager.DeleteMember(ctx, 2)
	assert.Nil(t, err)

	memberInfo, err := s.GetMemberOfResource(ctx, common.ResourceTemplate, fmt.Sprintf("%d", template.ID))
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
		&templatemodels.Template{})
	assert.Nil(t, err)

	ctx = context.Background()
	manager = managerparam.InitManager(db)
}
