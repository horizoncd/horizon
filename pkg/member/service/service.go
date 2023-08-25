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
	"errors"
	"fmt"
	"strconv"

	"github.com/horizoncd/horizon/core/common"
	herror "github.com/horizoncd/horizon/core/errors"
	applicationmanager "github.com/horizoncd/horizon/pkg/application/manager"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	clustermanager "github.com/horizoncd/horizon/pkg/cluster/manager"
	memberctx "github.com/horizoncd/horizon/pkg/context"
	perror "github.com/horizoncd/horizon/pkg/errors"
	groupmanager "github.com/horizoncd/horizon/pkg/group/manager"
	"github.com/horizoncd/horizon/pkg/member"
	"github.com/horizoncd/horizon/pkg/member/models"
	oauthmanager "github.com/horizoncd/horizon/pkg/oauth/manager"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	prmanager "github.com/horizoncd/horizon/pkg/pr/manager"
	roleservice "github.com/horizoncd/horizon/pkg/rbac/role"
	templatemanager "github.com/horizoncd/horizon/pkg/template/manager"
	templatereleasemanager "github.com/horizoncd/horizon/pkg/templaterelease/manager"
	usermanager "github.com/horizoncd/horizon/pkg/user/manager"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	webhookmanager "github.com/horizoncd/horizon/pkg/webhook/manager"
)

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/member/service/service_mock.go -package=mock_service
type Service interface {
	// CreateMember post a new member
	CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error)
	// GetMember return the current user member of direct or parent
	GetMember(ctx context.Context, memberID uint) (*models.Member, error)
	// UpdateMember update the member by the memberID
	UpdateMember(ctx context.Context, memberID uint, role string) (*models.Member, error)
	// RemoveMember Remove the member by the memberID
	RemoveMember(ctx context.Context, memberID uint) error
	// ListMember list all the member of the resource
	ListMember(ctx context.Context, resourceType string, resourceID uint) ([]models.Member, error)
	// GetMemberOfResource return the current user's role of the resource (member from direct or parent)
	GetMemberOfResource(ctx context.Context, resourceType string, resourceID string) (*models.Member, error)
	// IsYourPermissionHigher helps to check if your permission is higher then specified member
	RequirePermissionEqualOrHigher(ctx context.Context, role, resourceType string, resourceID uint) error
}

type service struct {
	memberManager             member.Manager
	groupManager              groupmanager.Manager
	applicationManager        applicationmanager.Manager
	applicationClusterManager clustermanager.Manager
	templateManager           templatemanager.Manager
	templateReleaseManager    templatereleasemanager.Manager
	prMgr                     *prmanager.PRManager
	roleService               roleservice.Service
	oauthManager              oauthmanager.Manager
	userManager               usermanager.Manager
	webhookManager            webhookmanager.Manager
}

func NewService(roleService roleservice.Service, oauthManager oauthmanager.Manager,
	manager *managerparam.Manager) Service {
	return &service{
		memberManager:             manager.MemberMgr,
		groupManager:              manager.GroupMgr,
		applicationManager:        manager.ApplicationMgr,
		applicationClusterManager: manager.ClusterMgr,
		prMgr:                     manager.PRMgr,
		templateReleaseManager:    manager.TemplateReleaseMgr,
		templateManager:           manager.TemplateMgr,
		roleService:               roleService,
		oauthManager:              oauthManager,
		userManager:               manager.UserMgr,
		webhookManager:            manager.WebhookMgr,
	}
}

func (s *service) RequirePermissionEqualOrHigher(ctx context.Context, role,
	resourceType string, resourceID uint) error {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	if currentUser.IsAdmin() {
		return nil
	}
	var userMemberInfo *models.Member
	userMemberInfo, err = s.getMember(ctx, resourceType,
		resourceID, models.MemberUser, currentUser.GetID())
	if err != nil {
		return err
	}
	if userMemberInfo == nil {
		return herror.ErrNoPrivilege
	}

	comResult, err := s.roleService.RoleCompare(ctx, userMemberInfo.Role, role)
	if err != nil {
		return err
	}
	if comResult == roleservice.RoleSmaller {
		return herror.ErrNoPrivilege
	}
	return nil
}

func (s *service) CreateMember(ctx context.Context, postMember PostMember) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. check exist
	memberItem, err := s.memberManager.Get(ctx, models.ResourceType(postMember.ResourceType), postMember.ResourceID,
		postMember.MemberType, postMember.MemberInfo)
	if err != nil {
		return nil, err
	}
	if memberItem != nil {
		// if member exist, try to update the member
		return s.UpdateMember(ctx, memberItem.ID, postMember.Role)
	}

	// 2. check if current user can create the role
	err = s.RequirePermissionEqualOrHigher(ctx, postMember.Role, postMember.ResourceType, postMember.ResourceID)
	if err != nil {
		return nil, err
	}

	// 3. do create  member
	member, err := ConvertPostMemberToMember(postMember, currentUser)
	if err != nil {
		return nil, err
	}
	return s.memberManager.Create(ctx, member)
}

func (s *service) getOauthAppMember(ctx context.Context, clientID string) (*models.Member, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	app, err := s.oauthManager.GetOAuthApp(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if !app.IsGroupOwnerType() {
		return nil, herror.ErrOAuthNotGroupOwnerType
	}
	return s.getMember(ctx, common.ResourceGroup, app.OwnerID, models.MemberUser, currentUser.GetID())
}

func (s *service) getPipelinerunMember(ctx context.Context, pipelinerunID uint) (*models.Member, error) {
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	pipeline, err := s.prMgr.PipelineRun.GetByID(ctx, pipelinerunID)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		msg := fmt.Sprintf("pipeline do not found, pipelineID = %d", pipelinerunID)
		log.Warningf(ctx, msg)
		return nil, herror.NewErrNotFound(herror.MemberInfoInDB, msg)
	}
	return s.getMember(ctx, common.ResourceCluster,
		pipeline.ClusterID, models.MemberUser, currentUser.GetID())
}

func (s *service) listPipelinerunMember(ctx context.Context, pipelinerunID uint) ([]models.Member, error) {
	pipeline, err := s.prMgr.PipelineRun.GetByID(ctx, pipelinerunID)
	if err != nil {
		return nil, err
	}
	if pipeline == nil {
		msg := fmt.Sprintf("pipeline do not found, pipelineID = %d", pipelinerunID)
		log.Warningf(ctx, msg)
		return nil, herror.NewErrNotFound(herror.MemberInfoInDB, msg)
	}
	return s.ListMember(ctx, common.ResourceApplication,
		pipeline.ClusterID)
}

func (s *service) listWebhookMember(ctx context.Context, id uint) ([]models.Member, error) {
	if id == 0 {
		return nil, nil
	}
	webhook, err := s.webhookManager.GetWebhook(ctx, id)
	if err != nil {
		return nil, err
	}
	switch webhook.ResourceType {
	case common.ResourceGroup, common.ResourceApplication, common.ResourceCluster:
		return s.ListMember(ctx, webhook.ResourceType,
			webhook.ResourceID)
	default:
		return nil, nil
	}
}

func (s *service) listWebhookLogMember(ctx context.Context, id uint) ([]models.Member, error) {
	if id == 0 {
		return nil, nil
	}
	webhookLog, err := s.webhookManager.GetWebhookLog(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.listWebhookMember(ctx, webhookLog.WebhookID)
}

func (s *service) GetMemberOfResource(ctx context.Context,
	resourceType string, resourceIDStr string) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}
	var memberInfo *models.Member
	if resourceType == common.ResourcePipelinerun {
		resourceID, _ := strconv.Atoi(resourceIDStr)
		memberInfo, err = s.getPipelinerunMember(ctx, uint(resourceID))
	} else if resourceType == common.ResourceOauthApps {
		memberInfo, err = s.getOauthAppMember(ctx, resourceIDStr)
	} else {
		resourceID, _ := strconv.Atoi(resourceIDStr)
		memberInfo, err = s.getMember(ctx, resourceType, uint(resourceID), models.MemberUser, currentUser.GetID())
	}
	if err != nil {
		return nil, err
	}
	if memberInfo == nil {
		defaultRole := s.roleService.GetDefaultRole(ctx)
		if nil != defaultRole {
			resourceID, _ := strconv.Atoi(resourceIDStr)
			memberInfo = &models.Member{
				MemberType:   models.MemberUser,
				Role:         defaultRole.Name,
				MemberNameID: currentUser.GetID(),
				ResourceType: models.ResourceType(resourceType),
				ResourceID:   uint(resourceID),
			}
		}
	}
	return memberInfo, nil
}

func (s *service) GetMember(ctx context.Context, memberID uint) (*models.Member, error) {
	return s.memberManager.GetByID(ctx, memberID)
}

func (s *service) RemoveMember(ctx context.Context, memberID uint) error {
	// 1. get member
	memberItem, err := s.memberManager.GetByID(ctx, memberID)
	if err != nil {
		return err
	}
	if memberItem == nil {
		return herror.NewErrNotFound(herror.MemberInfoInDB, fmt.Sprintf("member %d does not exist", memberID))
	}

	// 2. check if the grant current user can remove the member
	err = s.RequirePermissionEqualOrHigher(ctx, memberItem.Role, string(memberItem.ResourceType), memberItem.ResourceID)
	if err != nil {
		return err
	}

	// 3. check if common user
	user, err := s.userManager.GetUserByID(ctx, memberItem.MemberNameID)
	if err != nil {
		return err
	}
	if user.UserType != usermodels.UserTypeCommon {
		return perror.Wrapf(herror.ErrParamInvalid, "member of user type %d does not support updated", user.UserType)
	}

	return s.memberManager.DeleteMember(ctx, memberID)
}

func (s *service) UpdateMember(ctx context.Context, memberID uint, role string) (*models.Member, error) {
	// 1. get the member
	memberItem, err := s.memberManager.GetByID(ctx, memberID)
	if err != nil {
		return nil, err
	}
	if memberItem == nil {
		return nil, herror.NewErrNotFound(herror.MemberInfoInDB, fmt.Sprintf("member %d does not exist", memberID))
	}

	// 2. check if the current user have the permission to update the role
	err = s.RequirePermissionEqualOrHigher(ctx, memberItem.Role, string(memberItem.ResourceType),
		memberItem.ResourceID)
	if err != nil {
		return nil, err
	}
	err = s.RequirePermissionEqualOrHigher(ctx, role, string(memberItem.ResourceType), memberItem.ResourceID)
	if err != nil {
		return nil, err
	}

	// 3. check if common user
	user, err := s.userManager.GetUserByID(ctx, memberItem.MemberNameID)
	if err != nil {
		return nil, err
	}
	if user.UserType != usermodels.UserTypeCommon {
		return nil, perror.Wrapf(herror.ErrParamInvalid, "member of user type %d does not support updated", user.UserType)
	}

	// 4. update the role
	return s.memberManager.UpdateByID(ctx, memberItem.ID, role)
}

func (s *service) ListMember(ctx context.Context, resourceType string, resourceID uint) ([]models.Member, error) {
	// get all the members
	var allMembers []models.Member
	var err error
	switch resourceType {
	case common.ResourceGroup:
		allMembers, err = s.listGroupMembers(ctx, resourceID)
	case common.ResourceApplication:
		allMembers, err = s.listApplicationMembers(ctx, resourceID)
	case common.ResourceCluster:
		allMembers, err = s.listApplicationInstanceMembers(ctx, resourceID)
	case common.ResourceTemplate:
		allMembers, err = s.listTemplateMembers(ctx, resourceID)
	case common.ResourceTemplateRelease:
		allMembers, err = s.listReleaseMembers(ctx, resourceID)
	case common.ResourcePipelinerun:
		allMembers, err = s.listPipelinerunMember(ctx, resourceID)
	case common.ResourceWebhook:
		allMembers, err = s.listWebhookMember(ctx, resourceID)
	case common.ResourceWebhookLog:
		allMembers, err = s.listWebhookLogMember(ctx, resourceID)
	default:
		err = errors.New("unsupported resourceType")
	}
	if err != nil {
		return nil, err
	}
	return allMembers, nil
}

func DeduplicateMember(members []models.Member) []models.Member {
	// deduplicate by memberType, memberInfo
	memberMap := make(map[string]models.Member)

	var retMembers []models.Member
	for _, item := range members {
		key := strconv.Itoa(int(item.MemberType)) + "-" + strconv.FormatUint(uint64(item.MemberNameID), 10)
		_, ok := memberMap[key]
		if !ok {
			memberMap[key] = item
			retMembers = append(retMembers, item)
		}
	}
	return retMembers
}

// getMember return the direct member or member from the parent
func (s *service) getMember(ctx context.Context, resourceType string, resourceID uint,
	memberType models.MemberType, memberInfo uint) (*models.Member, error) {
	members, err := s.ListMember(ctx, resourceType, resourceID)
	if err != nil {
		return nil, err
	}
	for _, item := range members {
		if item.MemberType == memberType &&
			item.MemberNameID == memberInfo {
			return &item, nil
		}
	}
	return nil, nil
}

func (s *service) listGroupMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	var (
		retMembers []models.Member
		members    []models.Member
		err        error
	)

	onCondition, onConditionOK := ctx.Value(memberctx.MemberQueryOnCondition).(bool)
	if directMemberOnly, ok := ctx.Value(memberctx.MemberDirectMemberOnly).(bool); ok && directMemberOnly {
		if onConditionOK && onCondition {
			members, err = s.memberManager.ListDirectMemberOnCondition(ctx, models.TypeGroup, resourceID)
		} else {
			members, err = s.memberManager.ListDirectMember(ctx, models.TypeGroup, resourceID)
		}
		if err != nil {
			return nil, err
		}
		return DeduplicateMember(members), nil
	}

	// 1. list all the groups of the group
	if s.groupManager.IsRootGroup(ctx, resourceID) {
		// XXX: make sure only admins could access root group
		return []models.Member{}, nil
	}
	groupInfo, err := s.groupManager.GetByID(ctx, resourceID)
	if err != nil {
		return nil, err
	}
	groupIDs := groupmanager.FormatIDsFromTraversalIDs(groupInfo.TraversalIDs)

	// 2. get all the direct service of group
	for i := len(groupIDs) - 1; i >= 0; i-- {
		if onConditionOK && onCondition {
			members, err = s.memberManager.ListDirectMemberOnCondition(ctx, models.TypeGroup, groupIDs[i])
		} else {
			members, err = s.memberManager.ListDirectMember(ctx, models.TypeGroup, groupIDs[i])
		}
		if err != nil {
			return nil, err
		}
		retMembers = append(retMembers, members...)
	}

	return DeduplicateMember(retMembers), nil
}

func (s *service) listApplicationMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	var (
		retMembers []models.Member
		members    []models.Member
		err        error
	)
	if onCondition, ok := ctx.Value(memberctx.MemberQueryOnCondition).(bool); ok && onCondition {
		members, err = s.memberManager.ListDirectMemberOnCondition(ctx, models.TypeApplication, resourceID)
	} else {
		members, err = s.memberManager.ListDirectMember(ctx, models.TypeApplication, resourceID)
	}
	if err != nil {
		return nil, err
	}
	retMembers = append(retMembers, members...)

	if directMemberOnly, ok := ctx.Value(memberctx.MemberDirectMemberOnly).(bool); !ok || !directMemberOnly {
		// 1. query the application's service
		applicationInfo, err := s.applicationManager.GetByID(ctx, resourceID)
		if err != nil {
			return nil, err
		}

		// 2. query the group's service
		members, err = s.listGroupMembers(ctx, applicationInfo.GroupID)
		if err != nil {
			return nil, err
		}
		retMembers = append(retMembers, members...)
	}

	return DeduplicateMember(retMembers), nil
}

func (s *service) listApplicationInstanceMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	var (
		retMembers []models.Member
		err        error
	)

	var members []models.Member
	if onCondition, ok := ctx.Value(memberctx.MemberQueryOnCondition).(bool); ok && onCondition {
		members, err = s.memberManager.ListDirectMemberOnCondition(ctx, models.TypeApplicationCluster, resourceID)
	} else {
		members, err = s.memberManager.ListDirectMember(ctx, models.TypeApplicationCluster, resourceID)
	}

	if err != nil {
		return nil, err
	}
	retMembers = append(retMembers, members...)

	if directMemberOnly, ok := ctx.Value(memberctx.MemberDirectMemberOnly).(bool); !ok || !directMemberOnly {
		// 1. query the application cluster's members
		clusterInfo, err := s.applicationClusterManager.GetByID(ctx, resourceID)
		if err != nil {
			return nil, err
		}
		// 2. query the application's members (contains the group's members)
		members, err = s.listApplicationMembers(ctx, clusterInfo.ApplicationID)
		if err != nil {
			return nil, err
		}

		retMembers = append(retMembers, members...)
	}

	return DeduplicateMember(retMembers), nil
}

// createMemberDirect for unit test
func (s *service) createMemberDirect(ctx context.Context, postMember PostMember) (*models.Member, error) {
	var currentUser userauth.User
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// 1. convert service
	member, err := ConvertPostMemberToMember(postMember, currentUser)
	if err != nil {
		return nil, err
	}
	return s.memberManager.Create(ctx, member)
}

func (s *service) listTemplateMembers(ctx context.Context, resourceID uint) ([]models.Member, error) {
	var (
		retMembers []models.Member
		members    []models.Member
		err        error
	)
	if onCondition, ok := ctx.Value(memberctx.MemberQueryOnCondition).(bool); ok && onCondition {
		members, err = s.memberManager.ListDirectMemberOnCondition(ctx, models.TypeTemplate, resourceID)
	} else {
		members, err = s.memberManager.ListDirectMember(ctx, models.TypeTemplate, resourceID)
	}
	if err != nil {
		return nil, err
	}
	retMembers = append(retMembers, members...)

	if directMemberOnly, ok := ctx.Value(memberctx.MemberDirectMemberOnly).(bool); !ok || !directMemberOnly {
		// 1. query the application's service
		templateInfo, err := s.templateManager.GetByID(ctx, resourceID)
		if err != nil {
			return nil, err
		}

		// 2. query the group's service
		members, err = s.listGroupMembers(ctx, templateInfo.GroupID)
		if err != nil {
			return nil, err
		}
		retMembers = append(retMembers, members...)
	}

	return DeduplicateMember(retMembers), nil
}

func (s *service) listReleaseMembers(ctx context.Context, id uint) ([]models.Member, error) {
	tr, err := s.templateReleaseManager.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.listTemplateMembers(ctx, tr.Template)
}
