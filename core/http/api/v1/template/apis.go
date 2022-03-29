package template

import (
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	templatectl "g.hz.netease.com/horizon/core/controller/template"
	templateschematagctl "g.hz.netease.com/horizon/core/controller/templateschematag"

	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	// param
	_templateParam     = "template"
	_releaseParam      = "release"
	_resourceTypeQuery = "resourceType"
	_clusterIDQuery    = "clusterID"
)

type API struct {
	templateCtl templatectl.Controller
	tagCtl      templateschematagctl.Controller
}

func NewAPI(ctl templatectl.Controller, tagCtl templateschematagctl.Controller) *API {
	return &API{
		templateCtl: ctl,
		tagCtl:      tagCtl,
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
	params := make(map[string]string)

	// get the  params
	for key, values := range c.Request.URL.Query() {
		for _, value := range values {
			params[key] = value
		}
	}

	// if cluster id exist, get tags from cluster as param
	if c.Query(_resourceTypeQuery) == "cluster" {
		clusterIDStr := c.Query(_clusterIDQuery)
		if clusterIDStr != "" {
			clusterID, err := strconv.ParseUint(clusterIDStr, 10, 0)
			if err != nil {
				log.Info(c, "clusterID not found or invalid")
				response.AbortWithRequestError(c, common.InvalidRequestParam, err.Error())
				return
			}
			tags, err := a.tagCtl.List(c, uint(clusterID))
			if err != nil {
				log.Error(c, err.Error())
				response.AbortWithInternalError(c, err.Error())
			}
			for _, tag := range tags.Tags {
				params[tag.Key] = tag.Value
			}
		}
	}

	schema, err := a.templateCtl.GetTemplateSchema(c, t, r, params)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, schema)
}
