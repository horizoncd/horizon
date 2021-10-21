package template

import (
	templatectl "g.hz.netease.com/horizon/core/controller/template"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_templateParam = "template"
	_releaseParam  = "release"
)

type API struct {
	templateCtl templatectl.Controller
}

func NewAPI(ctl templatectl.Controller) *API {
	return &API{
		templateCtl: ctl,
	}
}

func (a *API) ListTemplate(c *gin.Context) {
	templates, err := a.templateCtl.ListTemplate(c)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, templates)
}

func (a *API) ListTemplateRelease(c *gin.Context) {
	t := c.Param(_templateParam)
	templateReleases, err := a.templateCtl.ListTemplateRelease(c, t)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, templateReleases)
}

func (a *API) GetTemplateSchema(c *gin.Context) {
	t := c.Param(_templateParam)
	r := c.Param(_releaseParam)
	// get template schema by templateName and releaseName
	schema, err := a.templateCtl.GetTemplateSchema(c, t, r)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, schema)
}
