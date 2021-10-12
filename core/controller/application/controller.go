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
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/jsonschema"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const _errCodeApplicationNotFound = errors.ErrorCode("ApplicationNotFound")

type Controller interface {
	// GetApplication get an application
	GetApplication(ctx context.Context, name string) (*GetApplicationResponse, error)
	// CreateApplication create an application
	CreateApplication(ctx context.Context, groupID uint, request *CreateApplicationRequest) error
	// UpdateApplication update an application
	UpdateApplication(ctx context.Context, name string, request *UpdateApplicationRequest) error
	// DeleteApplication delete an application by name
	DeleteApplication(ctx context.Context, name string) error
}

type controller struct {
	applicationGitRepo   gitrepo.ApplicationGitRepo
	templateSchemaGetter templateschema.Getter
	applicationMgr       manager.Manager
	groupMgr             groupmanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(applicationGitRepo gitrepo.ApplicationGitRepo) Controller {
	return &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateschema.Gtr,
		applicationMgr:       manager.Mgr,
		groupMgr:             groupmanager.Mgr,
	}
}

func (c *controller) GetApplication(ctx context.Context, name string) (_ *GetApplicationResponse, err error) {
	const op = "application controller: get application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get application jsonBlob in git repo
	pipelineJSONBlob, applicationJSONBlob, err := c.applicationGitRepo.GetApplication(ctx, name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. get application in db
	app, err := c.applicationMgr.GetByName(ctx, name)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if app == nil {
		return nil, errors.E(op, http.StatusNotFound, _errCodeApplicationNotFound)
	}
	return ofApplicationModel(app, pipelineJSONBlob, applicationJSONBlob), nil
}

func (c *controller) CreateApplication(ctx context.Context, groupID uint,
	request *CreateApplicationRequest) (err error) {
	const op = "application controller: create application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. validate
	if err := validateApplicationName(request.Name); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), err)
	}
	if err := c.validateBase(ctx, request.Base); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
	}

	// 2. check groups or applications with the same name exists
	groups, err := c.groupMgr.GetByNameOrPathUnderParent(ctx, request.Name, request.Name, groupID)
	if err != nil {
		return errors.E(op, err)
	}
	if len(groups) > 0 {
		return errors.E(op, http.StatusConflict,
			errors.ErrorCode(common.InvalidRequestBody),
			fmt.Sprintf("appliction name is in conflict with group under the same groupID: %v", groupID))
	}

	app, err := c.applicationMgr.GetByName(ctx, request.Name)
	if err != nil {
		return errors.E(op, err)
	}
	if app != nil {
		return errors.E(op, http.StatusConflict,
			errors.ErrorCode(common.InvalidRequestBody),
			fmt.Sprintf("application name: %v is already be taken", request.Name))
	}

	// 3. create application in git repo
	if err := c.applicationGitRepo.CreateApplication(ctx, request.Name,
		request.TemplateInput.Pipeline, request.TemplateInput.Application); err != nil {
		return errors.E(op, err)
	}

	// 4. create application in db
	applicationModel := request.toApplicationModel(groupID)
	applicationModel.CreatedBy = currentUser.GetName()
	applicationModel.UpdatedBy = currentUser.GetName()
	if _, err := c.applicationMgr.Create(ctx, applicationModel); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	return nil
}

func (c *controller) UpdateApplication(ctx context.Context, name string,
	request *UpdateApplicationRequest) (err error) {
	const op = "application controller: update application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. validate
	if err := c.validateBase(ctx, request.Base); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
	}

	// 2. update application in git repo
	if err := c.applicationGitRepo.UpdateApplication(ctx, name,
		request.TemplateInput.Pipeline, request.TemplateInput.Application); err != nil {
		return errors.E(op, err)
	}

	// 3. update application in db
	applicationModel := request.toApplicationModel()
	applicationModel.UpdatedBy = currentUser.GetName()
	if _, err := c.applicationMgr.UpdateByName(ctx, name, applicationModel); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	return nil
}

func (c *controller) DeleteApplication(ctx context.Context, name string) (err error) {
	const op = "application controller: delete application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get application in db
	app, err := c.applicationMgr.GetByName(ctx, name)
	if err != nil {
		return errors.E(op, err)
	}
	if app == nil {
		return errors.E(op, http.StatusNotFound, _errCodeApplicationNotFound)
	}

	// 2. delete application in git repo
	if err := c.applicationGitRepo.DeleteApplication(ctx, name, app.ID); err != nil {
		return errors.E(op, err)
	}

	// 2. delete application in db
	if err := c.applicationMgr.DeleteByName(ctx, name); err != nil {
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
