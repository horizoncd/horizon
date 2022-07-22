package application

import (
	"context"
	"fmt"
	"regexp"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	applicationmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/application/models"
	applicationservice "g.hz.netease.com/horizon/pkg/application/service"
	applicationregionmanager "g.hz.netease.com/horizon/pkg/applicationregion/manager"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	perror "g.hz.netease.com/horizon/pkg/errors"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/member"
	membermodels "g.hz.netease.com/horizon/pkg/member/models"
	"g.hz.netease.com/horizon/pkg/param"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	usersvc "g.hz.netease.com/horizon/pkg/user/service"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/jsonschema"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	// GetApplication get an application
	GetApplication(ctx context.Context, id uint) (*GetApplicationResponse, error)
	// CreateApplication create an application
	CreateApplication(ctx context.Context, groupID uint,
		request *CreateApplicationRequest) (*GetApplicationResponse, error)
	// UpdateApplication update an application
	UpdateApplication(ctx context.Context, id uint,
		request *UpdateApplicationRequest) (*GetApplicationResponse, error)
	// DeleteApplication delete an application by name
	DeleteApplication(ctx context.Context, id uint, hard bool) error
	// ListApplication list application by filter
	ListApplication(ctx context.Context, filter string, query q.Query) (int, []*ListApplicationResponse, error)
	// ListUserApplication list application that authorized to current user by filter
	ListUserApplication(ctx context.Context, filter string, query *q.Query) (int, []*ListApplicationResponse, error)
	// Transfer  try transfer application to another group
	Transfer(ctx context.Context, id uint, groupID uint) error
	GetSelectableRegionsByEnv(ctx context.Context, id uint, env string) (regionmodels.RegionParts, error)
}

type controller struct {
	applicationGitRepo   gitrepo.ApplicationGitRepo
	templateSchemaGetter templateschema.Getter
	applicationMgr       applicationmanager.Manager
	applicationSvc       applicationservice.Service
	groupMgr             groupmanager.Manager
	groupSvc             groupsvc.Service
	templateReleaseMgr   trmanager.Manager
	clusterMgr           clustermanager.Manager
	hook                 hook.Hook
	userSvc              usersvc.Service
	memberManager        member.Manager
	applicationRegionMgr applicationregionmanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		applicationGitRepo:   param.ApplicationGitRepo,
		templateSchemaGetter: param.TemplateSchemaGetter,
		applicationMgr:       param.ApplicationManager,
		applicationSvc:       param.ApplicationSvc,
		groupMgr:             param.GroupManager,
		groupSvc:             param.GroupSvc,
		templateReleaseMgr:   param.TemplateReleaseManager,
		clusterMgr:           param.ClusterMgr,
		hook:                 param.Hook,
		userSvc:              param.UserSvc,
		memberManager:        param.MemberManager,
		applicationRegionMgr: param.ApplicationRegionManager,
	}
}

func (c *controller) GetApplication(ctx context.Context, id uint) (_ *GetApplicationResponse, err error) {
	const op = "application controller: get application"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get application in db
	app, err := c.applicationMgr.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. get application jsonBlob in git repo
	pipelineJSONBlob, applicationJSONBlob, err := c.applicationGitRepo.GetApplication(ctx, app.Name)
	if err != nil {
		return nil, err
	}

	// 3. list template releases
	trs, err := c.templateReleaseMgr.ListByTemplateName(ctx, app.Template)
	if err != nil {
		return nil, err
	}

	// 4. get group full path
	group, err := c.groupSvc.GetChildByID(ctx, app.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, app.Name)
	return ofApplicationModel(app, fullPath, trs, pipelineJSONBlob, applicationJSONBlob), nil
}

func (c *controller) postHook(ctx context.Context, eventType hook.EventType, content interface{}) {
	if c.hook != nil {
		event := hook.Event{
			EventType: eventType,
			Event:     content,
		}
		go c.hook.Push(ctx, event)
	}
}

func (c *controller) CreateApplication(ctx context.Context, groupID uint,
	request *CreateApplicationRequest) (_ *GetApplicationResponse, err error) {
	const op = "application controller: create application"
	defer wlog.Start(ctx, op).StopPrint()

	extraMembers := request.ExtraMembers

	users := make([]string, 0, len(extraMembers))
	for member := range extraMembers {
		users = append(users, member)
	}

	// 1. validate
	err = c.userSvc.CheckUsersExists(ctx, users)
	if err != nil {
		return nil, err
	}

	if err := validateApplicationName(request.Name); err != nil {
		return nil, err
	}
	if err := c.validateCreate(request.Base); err != nil {
		return nil, err
	}
	if request.TemplateInput == nil {
		return nil, err
	}

	if err := c.validateTemplateInput(ctx, request.Template.Name,
		request.Template.Release, request.TemplateInput); err != nil {
		return nil, err
	}
	group, err := c.groupSvc.GetChildByID(ctx, groupID)
	if err != nil {
		return nil, err
	}

	// 2. check groups or applications with the same name exists
	groups, err := c.groupMgr.GetByNameOrPathUnderParent(ctx, request.Name, request.Name, groupID)
	if err != nil {
		return nil, err
	}
	if len(groups) > 0 {
		return nil, perror.Wrap(herrors.ErrNameConflict, "an group with the same name already exists")
	}

	appExistsInDB, err := c.applicationMgr.GetByName(ctx, request.Name)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return nil, err
		}
	}
	if appExistsInDB != nil {
		return nil, perror.Wrap(herrors.ErrNameConflict, "an application with the same name already exists, "+
			"please do not create it again")
	}

	// 3. create application in git repo
	if err := c.applicationGitRepo.CreateApplication(ctx, request.Name,
		request.TemplateInput.Pipeline, request.TemplateInput.Application); err != nil {
		return nil, err
	}

	// 4. create application in db
	applicationModel := request.toApplicationModel(groupID)
	applicationModel, err = c.applicationMgr.Create(ctx, applicationModel, extraMembers)
	if err != nil {
		return nil, err
	}

	// 5. get fullPath
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, applicationModel.Name)

	// 6. list template release
	trs, err := c.templateReleaseMgr.ListByTemplateName(ctx, applicationModel.Template)
	if err != nil {
		return nil, err
	}

	ret := ofApplicationModel(applicationModel, fullPath, trs,
		request.TemplateInput.Pipeline, request.TemplateInput.Application)

	// 7. post hook
	c.postHook(ctx, hook.CreateApplication, ret)

	return ret, nil
}

func (c *controller) UpdateApplication(ctx context.Context, id uint,
	request *UpdateApplicationRequest) (_ *GetApplicationResponse, err error) {
	const op = "application controller: update application"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get application in db
	appExistsInDB, err := c.applicationMgr.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// 2. validate
	if err := c.validateUpdate(request.Base); err != nil {
		return nil, err
	}

	// 3. if templateInput is not empty, validate and update in git repo
	if request.TemplateInput != nil {
		var template, templateRelease string
		if request.Template != nil {
			template = request.Template.Name
			templateRelease = request.Template.Release
		} else {
			template = appExistsInDB.Template
			templateRelease = appExistsInDB.TemplateRelease
		}
		if err := c.validateTemplateInput(ctx, template, templateRelease, request.TemplateInput); err != nil {
			return nil, err
		}
		if err := c.applicationGitRepo.UpdateApplication(ctx, appExistsInDB.Name,
			request.TemplateInput.Pipeline, request.TemplateInput.Application); err != nil {
			return nil, errors.E(op, err)
		}
	}

	// 4. update application in db
	applicationModel := request.toApplicationModel(appExistsInDB)
	applicationModel, err = c.applicationMgr.UpdateByID(ctx, id, applicationModel)
	if err != nil {
		return nil, err
	}

	// 5. get fullPath
	group, err := c.groupSvc.GetChildByID(ctx, appExistsInDB.GroupID)
	if err != nil {
		return nil, err
	}
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, appExistsInDB.Name)

	// 6. list template release
	trs, err := c.templateReleaseMgr.ListByTemplateName(ctx, appExistsInDB.Template)
	if err != nil {
		return nil, err
	}

	return ofApplicationModel(applicationModel, fullPath, trs,
		request.TemplateInput.Pipeline, request.TemplateInput.Application), nil
}

func (c *controller) DeleteApplication(ctx context.Context, id uint, hard bool) (err error) {
	const op = "application controller: delete application"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get application in db
	app, err := c.applicationMgr.GetByID(ctx, id)
	if err != nil {
		return err
	}

	clusters, err := c.clusterMgr.ListByApplicationID(ctx, id)
	if err != nil {
		return err
	}

	if len(clusters) > 0 {
		return perror.Wrap(herrors.ErrParamInvalid, "this application cannot be deleted "+
			"because there are clusters under this application.")
	}

	// 2. delete application in git repo
	if hard {
		// delete region config
		if err := c.applicationRegionMgr.UpsertByApplicationID(ctx, app.ID, nil); err != nil {
			return err
		}
		// delete git repo
		if err := c.applicationGitRepo.HardDeleteApplication(ctx, app.Name); err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
				return err
			}
		}
	} else {
		if err := c.applicationGitRepo.DeleteApplication(ctx, app.Name, app.ID); err != nil {
			return err
		}
	}

	// 3. delete application in db
	if err := c.applicationMgr.DeleteByID(ctx, id); err != nil {
		return err
	}

	// 4. post hook
	c.postHook(ctx, hook.DeleteApplication, app.Name)

	return nil
}

func (c *controller) Transfer(ctx context.Context, id uint, groupID uint) error {
	const op = "application controller: transfer application"
	defer wlog.Start(ctx, op).StopPrint()

	group, err := c.groupMgr.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	return c.applicationMgr.Transfer(ctx, id, group.ID)
}

func (c *controller) validateCreate(b Base) error {
	if err := validatePriority(b.Priority); err != nil {
		return err
	}
	if b.Template == nil {
		return perror.Wrap(herrors.ErrParamInvalid, "template cannot be empty")
	}
	return validateGit(b)
}

func (c *controller) validateUpdate(b Base) error {
	if b.Priority != "" {
		if err := validatePriority(b.Priority); err != nil {
			return err
		}
	}
	return validateGit(b)
}

func validateGit(b Base) error {
	if b.Git != nil && b.Git.URL != "" {
		re := `^ssh://.+[.]git$`
		pattern := regexp.MustCompile(re)
		if !pattern.MatchString(b.Git.URL) {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("invalid git url, should satisfies the pattern %v", re))
		}
	}
	return nil
}

// validateTemplateInput validate templateInput is valid for template schema
func (c *controller) validateTemplateInput(ctx context.Context,
	template, release string, templateInput *TemplateInput) error {
	schema, err := c.templateSchemaGetter.GetTemplateSchema(ctx, template, release, nil)
	if err != nil {
		return err
	}
	if err := jsonschema.Validate(schema.Application.JSONSchema, templateInput.Application, false); err != nil {
		return err
	}
	return jsonschema.Validate(schema.Pipeline.JSONSchema, templateInput.Pipeline, true)
}

// validatePriority validate priority
func validatePriority(priority string) error {
	switch models.Priority(priority) {
	case models.P0, models.P1, models.P2, models.P3:
	default:
		return perror.Wrap(herrors.ErrParamInvalid, "invalid priority")
	}
	return nil
}

// validateApplicationName validate application name
// 1. name length must be less than 40
// 2. name must match pattern ^(([a-z][-a-z0-9]*)?[a-z0-9])?$
func validateApplicationName(name string) error {
	if len(name) == 0 {
		return perror.Wrap(herrors.ErrParamInvalid, "name cannot be empty")
	}

	if len(name) > 40 {
		return perror.Wrap(herrors.ErrParamInvalid, "name must not exceed 40 characters")
	}

	// cannot start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		return perror.Wrap(herrors.ErrParamInvalid, "name cannot start with a digit")
	}

	pattern := `^(([a-z][-a-z0-9]*)?[a-z0-9])?$`
	r := regexp.MustCompile(pattern)
	if !r.MatchString(name) {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid application name, regex used for validation is %v", pattern))
	}

	return nil
}

func (c *controller) ListApplication(ctx context.Context, filter string, query q.Query) (count int,
	listApplicationResp []*ListApplicationResponse, err error) {
	const op = "application controller: list application"
	defer wlog.Start(ctx, op).StopPrint()

	listApplicationResp = []*ListApplicationResponse{}
	// 1. get application in db
	count, applications, err := c.applicationMgr.GetByNameFuzzilyByPagination(ctx, filter, query)
	if err != nil {
		return count, nil, err
	}

	// 2. get groups for full path, full name
	var groupIDs []uint
	for _, application := range applications {
		groupIDs = append(groupIDs, application.GroupID)
	}
	groupMap, err := c.groupSvc.GetChildrenByIDs(ctx, groupIDs)
	if err != nil {
		return count, nil, err
	}

	// 3. convert models.Application to ListApplicationResponse
	for _, application := range applications {
		group, exist := groupMap[application.GroupID]
		if !exist {
			continue
		}
		fullPath := fmt.Sprintf("%v/%v", group.FullPath, application.Name)
		fullName := fmt.Sprintf("%v/%v", group.FullName, application.Name)
		listApplicationResp = append(
			listApplicationResp,
			&ListApplicationResponse{
				FullName:  fullName,
				FullPath:  fullPath,
				Name:      application.Name,
				GroupID:   application.GroupID,
				ID:        application.ID,
				CreatedAt: application.CreatedAt,
				UpdatedAt: application.UpdatedAt,
			},
		)
	}

	return count, listApplicationResp, nil
}

func (c *controller) ListUserApplication(ctx context.Context,
	filter string, query *q.Query) (int, []*ListApplicationResponse, error) {
	// get current user
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return 0, nil, perror.WithMessage(err, "no user in context")
	}

	// get groups authorized to current user
	groupIDs, err := c.memberManager.ListResourceOfMemberInfo(ctx, membermodels.TypeGroup, currentUser.GetID())
	if err != nil {
		return 0, nil,
			perror.WithMessage(err, "failed to list group resource of current user")
	}

	// get these groups' subGroups
	subGroups, err := c.groupMgr.GetSubGroupsByGroupIDs(ctx, groupIDs)
	if err != nil {
		return 0, nil, perror.WithMessage(err, "failed to get groups")
	}

	subGroupIDs := make([]uint, 0)
	for _, group := range subGroups {
		subGroupIDs = append(subGroupIDs, group.ID)
	}

	count, applications, err := c.applicationMgr.ListUserAuthorizedByNameFuzzily(ctx,
		filter, subGroupIDs, currentUser.GetID(), query)
	if err != nil {
		return 0, nil, perror.WithMessage(err, "failed to list user applications")
	}

	// get groups for full path, full name
	var applicationGroupIDs []uint
	for _, application := range applications {
		applicationGroupIDs = append(applicationGroupIDs, application.GroupID)
	}
	groupMap, err := c.groupSvc.GetChildrenByIDs(ctx, applicationGroupIDs)
	if err != nil {
		return count, nil, err
	}

	// convert models.Application to ListApplicationResponse
	listApplicationResp := make([]*ListApplicationResponse, 0)
	for _, application := range applications {
		group, exist := groupMap[application.GroupID]
		if !exist {
			continue
		}
		fullPath := fmt.Sprintf("%v/%v", group.FullPath, application.Name)
		fullName := fmt.Sprintf("%v/%v", group.FullName, application.Name)
		listApplicationResp = append(
			listApplicationResp,
			&ListApplicationResponse{
				FullName:  fullName,
				FullPath:  fullPath,
				Name:      application.Name,
				GroupID:   application.GroupID,
				ID:        application.ID,
				CreatedAt: application.CreatedAt,
				UpdatedAt: application.UpdatedAt,
			},
		)
	}

	return count, listApplicationResp, nil
}

func (c *controller) GetSelectableRegionsByEnv(ctx context.Context, id uint, env string) (
	regionmodels.RegionParts, error) {
	application, err := c.applicationMgr.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	selectableRegionsByEnv, err := c.groupMgr.GetSelectableRegionsByEnv(ctx, application.GroupID, env)
	if err != nil {
		return nil, err
	}

	// set isDefault field
	applicationRegion, err := c.applicationRegionMgr.ListByEnvApplicationID(ctx, env, id)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return nil, err
		}
	}
	if applicationRegion != nil {
		for _, regionPart := range selectableRegionsByEnv {
			regionPart.IsDefault = regionPart.Name == applicationRegion.RegionName
		}
	}

	return selectableRegionsByEnv, nil
}
