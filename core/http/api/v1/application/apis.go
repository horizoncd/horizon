package application

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/controller/template"
	"g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/server/response"
	"g.hz.netease.com/horizon/util/jsonschema"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_applicationParam = "application"
)

type API struct {
	templateCtl template.Controller
}

func (a *API) Get(c *gin.Context) {

}

func (a *API) Create(c *gin.Context) {
	var applicationModel *CreateApplication
	if err := c.ShouldBindJSON(&applicationModel); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody,
			fmt.Sprintf("request body is invalid, err: %v", err))
		return
	}
	if err := a.validate(c, applicationModel); err != nil {
		response.AbortWithRequestError(c, common.InvalidRequestBody, err.Error())
		return
	}
}

func (a *API) Update(c *gin.Context) {

}

func (a *API) validate(c *gin.Context, applicationModel *CreateApplication) error {
	t := applicationModel.Template
	tInput := applicationModel.TemplateInput
	if err := a.validatePriority(applicationModel.Priority); err != nil {
		return err
	}
	if err := a.validateTemplateInput(c, t.Name, t.Release, tInput); err != nil {
		return err
	}
	return nil
}

// validateTemplateInput validate templateInput is valid for template schema
func (a *API) validateTemplateInput(ctx context.Context,
	template, release string, templateInput map[string]interface{}) error {
	schema, err := a.templateCtl.GetTemplateSchema(ctx, template, release)
	if err != nil {
		return err
	}
	return jsonschema.Validate(schema, templateInput)
}

func (*API) validatePriority(priority string) error {
	switch models.Priority(priority) {
	case models.P0, models.P1, models.P2, models.P3:
	default:
		return fmt.Errorf("invalid priority")
	}
	return nil
}
