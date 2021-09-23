package template

import (
	"encoding/json"

	"g.hz.netease.com/horizon/controller/gitlab"
	"g.hz.netease.com/horizon/pkg/template"
	"g.hz.netease.com/horizon/pkg/templaterelease"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const (
	templateParam = "template"
	releaseParam  = "release"
)

type API struct {
	templateMgr        template.Manager
	templateReleaseMgr templaterelease.Manager
	gitlabCtl          gitlab.Controller
}

func NewAPI() *API {
	return &API{
		templateMgr:        template.Mgr,
		templateReleaseMgr: templaterelease.Mgr,
		gitlabCtl:          gitlab.Ctl,
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
	t := c.Param(templateParam)
	templates, err := a.templateReleaseMgr.ListByTemplateName(c, t)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, toReleases(templates))
}

const (
	schemaPath = "schema/input.schema.json"
)

func (a *API) GetTemplateSchema(c *gin.Context) {
	t := c.Param(templateParam)
	r := c.Param(releaseParam)
	tr, err := a.templateReleaseMgr.GetByTemplateNameAndRelease(c, t, r)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	gitlabLib, err := a.gitlabCtl.GetByName(c, tr.GitlabName)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	b, err := gitlabLib.GetFile(c, tr.GitlabProject, tr.Name, schemaPath)
	if err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	var jsonSchema map[string]interface{}
	if err := json.Unmarshal(b, &jsonSchema); err != nil {
		response.AbortWithInternalError(c, err.Error())
		return
	}
	response.SuccessWithData(c, jsonSchema)
}
