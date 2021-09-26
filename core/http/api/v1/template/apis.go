package template

import (
	"encoding/json"

	templatectl "g.hz.netease.com/horizon/controller/template"
	"g.hz.netease.com/horizon/pkg/template"
	"g.hz.netease.com/horizon/pkg/templaterelease"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_templateParam = "template"
	_releaseParam  = "release"
)

type API struct {
	templateMgr        template.Manager
	templateReleaseMgr templaterelease.Manager
	templateCtl        templatectl.Controller
}

func NewAPI() *API {
	return &API{
		templateMgr:        template.Mgr,
		templateReleaseMgr: templaterelease.Mgr,
		templateCtl:        templatectl.Ctl,
	}
}

func (a *API) ListTemplate(c *gin.Context) {
	templates, err := a.templateMgr.List(c)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, toTemplates(templates))
}

func (a *API) ListTemplateRelease(c *gin.Context) {
	t := c.Param(_templateParam)
	templates, err := a.templateReleaseMgr.ListByTemplateName(c, t)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, toReleases(templates))
}

func (a *API) GetTemplateSchema(c *gin.Context) {
	t := c.Param(_templateParam)
	r := c.Param(_releaseParam)
	// get template schema by templateName and releaseName
	b, err := a.templateCtl.GetTemplateSchema(c, t, r)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	// convert template json schema to map[string]interface{}
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal(b, &jsonSchema); err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, jsonSchema)
}
