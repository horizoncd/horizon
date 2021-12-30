package envtemplate

import (
	"context"
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/pkg/application/gitrepo"
	"g.hz.netease.com/horizon/pkg/application/manager"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	templateschema "g.hz.netease.com/horizon/pkg/templaterelease/schema"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/jsonschema"
	"g.hz.netease.com/horizon/pkg/util/sets"
)

type Controller interface {
	UpdateEnvTemplate(ctx context.Context, applicationID uint, env string, r *UpdateEnvTemplateRequest) error
	GetEnvTemplate(ctx context.Context, applicationID uint, env string) (*GetEnvTemplateResponse, error)
}

type controller struct {
	applicationGitRepo   gitrepo.ApplicationGitRepo
	templateSchemaGetter templateschema.Getter
	applicationMgr       manager.Manager
	envMgr               envmanager.Manager
}

func NewController(applicationGitRepo gitrepo.ApplicationGitRepo,
	templateSchemaGetter templateschema.Getter) Controller {
	return &controller{
		applicationGitRepo:   applicationGitRepo,
		templateSchemaGetter: templateSchemaGetter,
		applicationMgr:       manager.Mgr,
		envMgr:               envmanager.Mgr,
	}
}

func (c *controller) UpdateEnvTemplate(ctx context.Context,
	applicationID uint, env string, r *UpdateEnvTemplateRequest) error {
	const op = "env template controller: update env templates"

	// 1. get application
	application, err := c.applicationMgr.GetByID(ctx, applicationID)
	if err != nil {
		return errors.E(op, err)
	}

	// 2. validate schema
	schema, err := c.templateSchemaGetter.GetTemplateSchema(ctx, application.Template, application.TemplateRelease, nil)
	if err != nil {
		return errors.E(op, http.StatusBadRequest, err)
	}
	if err := jsonschema.Validate(schema.Application.JSONSchema, r.Application); err != nil {
		return errors.E(op, http.StatusBadRequest, err)
	}
	if err := jsonschema.Validate(schema.Pipeline.JSONSchema, r.Pipeline); err != nil {
		return errors.E(op, http.StatusBadRequest, err)
	}

	// 3.1 update application's git repo if env is empty
	if env == "" {
		if err := c.applicationGitRepo.UpdateApplication(ctx, application.Name,
			r.Pipeline, r.Application); err != nil {
			return errors.E(op, err)
		}
		return nil
	}

	// 3.2 check env exists
	if err := c.checkEnvExists(ctx, env); err != nil {
		return errors.E(op, err)
	}
	// 4. update application env template in git repo
	return c.applicationGitRepo.UpdateApplicationEnvTemplate(ctx, application.Name, env, r.Pipeline, r.Application)
}

func (c *controller) GetEnvTemplate(ctx context.Context,
	applicationID uint, env string) (*GetEnvTemplateResponse, error) {
	const op = "env template controller: get env templates"

	// 1. get application
	application, err := c.applicationMgr.GetByID(ctx, applicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	var pipelineJSONBlob, applicationJSONBlob map[string]interface{}
	// 2.1 get application's git repo if env is empty
	if env == "" {
		pipelineJSONBlob, applicationJSONBlob, err = c.applicationGitRepo.GetApplication(ctx, application.Name)
	} else {
		// 2.2 check env exists
		if err := c.checkEnvExists(ctx, env); err != nil {
			return nil, errors.E(op, err)
		}
		// 3. get application env template
		pipelineJSONBlob, applicationJSONBlob, err = c.applicationGitRepo.GetApplicationEnvTemplate(ctx,
			application.Name, env)
	}

	if err != nil {
		return nil, errors.E(op, err)
	}
	return &GetEnvTemplateResponse{
		EnvTemplate: &EnvTemplate{
			Application: applicationJSONBlob,
			Pipeline:    pipelineJSONBlob,
		},
	}, nil
}

func (c *controller) checkEnvExists(ctx context.Context, envName string) error {
	const op = "env template controller: check env exists"

	envs, err := c.envMgr.ListAllEnvironment(ctx)
	if err != nil {
		return err
	}
	envSet := sets.NewString()
	for _, env := range envs {
		envSet.Insert(env.Name)
	}
	if !envSet.Has(envName) {
		return errors.E(op, http.StatusNotFound, fmt.Sprintf("environment %s is not exists", envName))
	}
	return nil
}
