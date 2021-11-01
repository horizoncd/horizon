package application

import (
	"context"
	"fmt"
	"net/http"
	"regexp"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	"g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/application/models"
	groupmanager "g.hz.netease.com/horizon/pkg/group/manager"
	groupsvc "g.hz.netease.com/horizon/pkg/group/service"
	trmanager "g.hz.netease.com/horizon/pkg/templaterelease/manager"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
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
	DeleteApplication(ctx context.Context, id uint) error
}

type controller struct {
	applicationGitRepo   gitrepo.ApplicationGitRepo
	templateSchemaGetter templateschema.Getter
	applicationMgr       manager.Manager
	groupMgr             groupmanager.Manager
	groupSvc             groupsvc.Service
	templateReleaseMgr   trmanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(applicationGitRepo gitrepo.ApplicationGitRepo,
	templateSchemaGetter templateschema.Getter) Controller {
	return &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateSchemaGetter,
		applicationMgr:       manager.Mgr,
		groupMgr:             groupmanager.Mgr,
		groupSvc:             groupsvc.Svc,
		templateReleaseMgr:   trmanager.Mgr,
	}
}

func (c *controller) GetApplication(ctx context.Context, id uint) (_ *GetApplicationResponse, err error) {
	const op = "application controller: get application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get application in db
	app, err := c.applicationMgr.GetByID(ctx, id)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. get application jsonBlob in git repo
	pipelineJSONBlob, applicationJSONBlob, err := c.applicationGitRepo.GetApplication(ctx, app.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. list template releases
	trs, err := c.templateReleaseMgr.ListByTemplateName(ctx, app.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 4. get group full path
	group, err := c.groupSvc.GetChildByID(ctx, app.GroupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, app.Name)
	return ofApplicationModel(app, fullPath, trs, pipelineJSONBlob, applicationJSONBlob), nil
}

func (c *controller) CreateApplication(ctx context.Context, groupID uint,
	request *CreateApplicationRequest) (_ *GetApplicationResponse, err error) {
	const op = "application controller: create application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. validate
	if err := validateApplicationName(request.Name); err != nil {
		return nil, errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}
	if err := c.validateBase(ctx, request.Base); err != nil {
		return nil, errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
	}

	// 2. check groups or applications with the same name exists
	groups, err := c.groupMgr.GetByNameOrPathUnderParent(ctx, request.Name, request.Name, groupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if len(groups) > 0 {
		return nil, errors.E(op, http.StatusConflict,
			errors.ErrorCode(common.InvalidRequestBody),
			fmt.Sprintf("appliction name is in conflict with group under the same groupID: %v", groupID))
	}

	appExistsInDB, err := c.applicationMgr.GetByName(ctx, request.Name)
	if err != nil {
		if errors.Status(err) != http.StatusNotFound {
			return nil, errors.E(op, err)
		}
	}
	if appExistsInDB != nil {
		return nil, errors.E(op, http.StatusConflict,
			errors.ErrorCode(common.InvalidRequestBody),
			fmt.Sprintf("application name: %v is already be taken", request.Name))
	}

	// 3. create application in git repo
	if err := c.applicationGitRepo.CreateApplication(ctx, request.Name,
		request.TemplateInput.Pipeline, request.TemplateInput.Application); err != nil {
		return nil, errors.E(op, err)
	}

	// 4. create application in db
	applicationModel := request.toApplicationModel(groupID)
	applicationModel.CreatedBy = currentUser.GetID()
	applicationModel.UpdatedBy = currentUser.GetID()
	applicationModel, err = c.applicationMgr.Create(ctx, applicationModel)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 5. get fullPath
	group, err := c.groupSvc.GetChildByID(ctx, groupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, applicationModel.Name)

	// 6. list template release
	trs, err := c.templateReleaseMgr.ListByTemplateName(ctx, applicationModel.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ofApplicationModel(applicationModel, fullPath, trs,
		request.TemplateInput.Pipeline, request.TemplateInput.Application), nil
}

func (c *controller) UpdateApplication(ctx context.Context, id uint,
	request *UpdateApplicationRequest) (_ *GetApplicationResponse, err error) {
	const op = "application controller: update application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. get application in db
	appExistsInDB, err := c.applicationMgr.GetByID(ctx, id)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. validate
	if err := c.validateBase(ctx, request.Base); err != nil {
		return nil, errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
	}

	// 3. update application in git repo
	if err := c.applicationGitRepo.UpdateApplication(ctx, appExistsInDB.Name,
		request.TemplateInput.Pipeline, request.TemplateInput.Application); err != nil {
		return nil, errors.E(op, err)
	}

	// 4. update application in db
	applicationModel := request.toApplicationModel()
	applicationModel.UpdatedBy = currentUser.GetID()
	applicationModel, err = c.applicationMgr.UpdateByID(ctx, id, applicationModel)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 5. get fullPath
	group, err := c.groupSvc.GetChildByID(ctx, appExistsInDB.GroupID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	fullPath := fmt.Sprintf("%v/%v", group.FullPath, appExistsInDB.Name)

	// 6. list template release
	trs, err := c.templateReleaseMgr.ListByTemplateName(ctx, appExistsInDB.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return ofApplicationModel(applicationModel, fullPath, trs,
		request.TemplateInput.Pipeline, request.TemplateInput.Application), nil
}

func (c *controller) DeleteApplication(ctx context.Context, id uint) (err error) {
	const op = "application controller: delete application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get application in db
	app, err := c.applicationMgr.GetByID(ctx, id)
	if err != nil {
		return errors.E(op, err)
	}

	// 2. delete application in git repo
	if err := c.applicationGitRepo.DeleteApplication(ctx, app.Name, app.ID); err != nil {
		return errors.E(op, err)
	}

	// 3. delete application in db
	if err := c.applicationMgr.DeleteByID(ctx, id); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	return nil
}

func (c *controller) validateBase(ctx context.Context, b Base) error {
	t := b.Template
	tInput := b.TemplateInput
	if err := validatePriority(b.Priority); err != nil {
		return err
	}
	if err := c.validateTemplateInput(ctx, t.Name, t.Release, tInput); err != nil {
		return err
	}
	return nil
}

// validateTemplateInput validate templateInput is valid for template schema
func (c *controller) validateTemplateInput(ctx context.Context,
	template, release string, templateInput *TemplateInput) error {
	schema, err := c.templateSchemaGetter.GetTemplateSchema(ctx, template, release)
	if err != nil {
		return err
	}
	if err := jsonschema.Validate(schema.Application.JSONSchema, templateInput.Application); err != nil {
		return err
	}
	return jsonschema.Validate(schema.Pipeline.JSONSchema, templateInput.Pipeline)
}

// validatePriority validate priority
func validatePriority(priority string) error {
	switch models.Priority(priority) {
	case models.P0, models.P1, models.P2, models.P3:
	default:
		return fmt.Errorf("invalid priority")
	}
	return nil
}

// validateApplicationName validate application name
// 1. name length must be less than 40
// 2. name must match pattern ^(([a-z][-a-z0-9]*)?[a-z0-9])?$
func validateApplicationName(name string) error {
	if len(name) == 0 {
		return fmt.Errorf("name cannot be empty")
	}

	if len(name) > 40 {
		return fmt.Errorf("name must not exceed 40 characters")
	}

	// cannot start with a digit.
	if name[0] >= '0' && name[0] <= '9' {
		return fmt.Errorf("name cannot start with a digit")
	}

	pattern := `^(([a-z][-a-z0-9]*)?[a-z0-9])?$`
	r := regexp.MustCompile(pattern)
	if !r.MatchString(name) {
		return fmt.Errorf("invalid application name, regex used for validation is %v", pattern)
	}

	return nil
}