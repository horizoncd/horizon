package application

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"g.hz.netease.com/horizon/common"
	gitlabctl "g.hz.netease.com/horizon/controller/gitlab"
	"g.hz.netease.com/horizon/controller/template"
	"g.hz.netease.com/horizon/core/middleware/user"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	"g.hz.netease.com/horizon/pkg/application"
	"g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	"g.hz.netease.com/horizon/util/angular"
	"g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/jsonschema"
	"g.hz.netease.com/horizon/util/wlog"
)

const (
	_branchMaster = "master"

	_filePathTemplate = "template.json"
	_filePathPipeline = "pipeline.json"
)

const _errCodeApplicationNotFound = errors.ErrorCode("ApplicationNotFound")

type Controller interface {
	// GetApplication get an application
	GetApplication(ctx context.Context, name string) (*GetApplicationResponse, error)
	// CreateApplication create an application
	CreateApplication(ctx context.Context, request *CreateApplicationRequest) error
	// UpdateApplication update an application
	UpdateApplication(ctx context.Context, name string, request *UpdateApplicationRequest) error
	// DeleteApplication delete an application by name
	DeleteApplication(ctx context.Context, name string) error
}

type controller struct {
	gitlabConfig   gitlab.Config
	templateCtl    template.Controller
	gitlabCtl      gitlabctl.Controller
	applicationMgr application.Manager
}

var _ Controller = (*controller)(nil)

func NewController(gitlabConfig gitlab.Config) Controller {
	return &controller{
		gitlabConfig:   gitlabConfig,
		templateCtl:    template.Ctl,
		gitlabCtl:      gitlabctl.Ctl,
		applicationMgr: application.Mgr,
	}
}

func (c *controller) GetApplication(ctx context.Context, name string) (_ *GetApplicationResponse, err error) {
	const op = "application controller: get application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	var applicationConf = c.gitlabConfig.Application

	// 1. get gitlab lib instance
	gitlabLib, err := c.gitlabCtl.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 2. get template and pipeline from gitlab
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, name)
	var templateBytes, pipelineBytes []byte
	var err1, err2 error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		templateBytes, err1 = gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathTemplate)
	}()
	go func() {
		defer wg.Done()
		pipelineBytes, err2 = gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathPipeline)
	}()
	wg.Wait()

	if err1 != nil {
		return nil, errors.E(op, err1)
	}
	if err2 != nil {
		return nil, errors.E(op, err2)
	}

	var templateInput, pipelineInput map[string]interface{}
	if err := json.Unmarshal(templateBytes, &templateInput); err != nil {
		return nil, errors.E(op, err)
	}
	if err := json.Unmarshal(pipelineBytes, &pipelineInput); err != nil {
		return nil, errors.E(op, err)
	}
	app, err := c.applicationMgr.GetByName(ctx, name)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if app == nil {
		return nil, errors.E(op, http.StatusNotFound, _errCodeApplicationNotFound)
	}
	return ofApplicationModel(app, templateInput, pipelineInput), nil
}

func (c *controller) CreateApplication(ctx context.Context, request *CreateApplicationRequest) (err error) {
	const op = "application controller: create application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	var applicationConf = c.gitlabConfig.Application
	// 1. validate
	if err := c.validate(ctx, request.Base); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
	}

	// 2. get gitlab lib instance
	gitlabLib, err := c.gitlabCtl.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 3. create application project in gitlab
	if _, err := gitlabLib.CreateProject(ctx, request.Name, applicationConf.Parent.ID); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 4. write files to gitlab
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, request.Name)
	templateJson, err := json.MarshalIndent(request.TemplateInput, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	pipelineJson, err := json.MarshalIndent(request.PipelineInput, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileCreate,
			FilePath: _filePathTemplate,
			Content:  string(templateJson),
		}, {
			Action:   gitlablib.FileCreate,
			FilePath: _filePathPipeline,
			Content:  string(pipelineJson),
		},
	}

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.Name,
		Action:      "create application",
		Application: angular.StringPtr(request.Name),
	}, request)

	if _, err := gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 5. write to db
	applicationModel := request.toApplicationModel()
	applicationModel.CreatedBy = currentUser.Name
	applicationModel.UpdatedBy = currentUser.Name
	if _, err := c.applicationMgr.Create(ctx, applicationModel); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	return nil
}

func (c *controller) UpdateApplication(ctx context.Context, name string, request *UpdateApplicationRequest) (err error) {
	const op = "application controller: create application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	var applicationConf = c.gitlabConfig.Application
	// 1. validate
	if err := c.validate(ctx, request.Base); err != nil {
		return errors.E(op, http.StatusBadRequest,
			errors.ErrorCode(common.InvalidRequestBody), fmt.Sprintf("request body validate err: %v", err))
	}

	// 2. get gitlab lib instance
	gitlabLib, err := c.gitlabCtl.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 3. write files to gitlab
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, name)
	templateJson, err := json.MarshalIndent(request.TemplateInput, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	pipelineJson, err := json.MarshalIndent(request.PipelineInput, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathTemplate,
			Content:  string(templateJson),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathPipeline,
			Content:  string(pipelineJson),
		},
	}

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.Name,
		Action:      "update application",
		Application: angular.StringPtr(name),
	}, request)

	if _, err := gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 5. write to db
	applicationModel := request.toApplicationModel()
	applicationModel.UpdatedBy = currentUser.Name
	if _, err := c.applicationMgr.UpdateByName(ctx, name, applicationModel); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	return nil
}

func (c *controller) DeleteApplication(ctx context.Context, name string) (err error) {
	const op = "application controller: delete application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	var applicationConf = c.gitlabConfig.Application

	// 1. get gitlab lib instance
	gitlabLib, err := c.gitlabCtl.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 2. delete gitlab project
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, name)
	if err := gitlabLib.DeleteProject(ctx, pid); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 3. delete application in db
	if err := c.applicationMgr.DeleteByName(ctx, name); err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	return nil
}

func (c *controller) validate(ctx context.Context, b Base) error {
	t := b.Template
	tInput := b.TemplateInput
	if err := c.validatePriority(b.Priority); err != nil {
		return err
	}
	if err := c.validateTemplateInput(ctx, t.Name, t.Release, tInput); err != nil {
		return err
	}
	return nil
}

// validateTemplateInput validate templateInput is valid for template schema
func (c *controller) validateTemplateInput(ctx context.Context,
	template, release string, templateInput map[string]interface{}) error {
	schema, err := c.templateCtl.GetTemplateSchema(ctx, template, release)
	if err != nil {
		return err
	}
	return jsonschema.Validate(schema, templateInput)
}

// validatePriority validate priority
func (c *controller) validatePriority(priority string) error {
	switch models.Priority(priority) {
	case models.P0, models.P1, models.P2, models.P3:
	default:
		return fmt.Errorf("invalid priority")
	}
	return nil
}
