package template

import (
	"context"
	"fmt"
	"strconv"

	tplctx "github.com/horizoncd/horizon/pkg/context"
	"github.com/horizoncd/horizon/pkg/core/common"
	templatectl "github.com/horizoncd/horizon/pkg/core/controller/template"
	templateschematagctl "github.com/horizoncd/horizon/pkg/core/controller/templateschematag"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const (
	// param
	_templateParam = "templateID"
	_releaseParam  = "releaseID"
	_groupParam    = "groupID"

	// query
	_resourceTypeQuery = "resourceType"
	_clusterIDQuery    = "clusterID"
	_withFullPath      = "fullpath"
	_withReleases      = "withReleases"
	_listRecursively   = "recursive"
	_listSelfOnly      = "selfOnly"
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
	withFullPathStr := c.Query(_withFullPath)
	withFullPath, err := strconv.ParseBool(withFullPathStr)
	var ctx context.Context = c
	if err == nil {
		ctx = context.WithValue(ctx, tplctx.TemplateWithFullPath, withFullPath)
	}

	listSelfOnlyStr := c.Query(_listSelfOnly)
	listSelfOnly, err := strconv.ParseBool(listSelfOnlyStr)
	if err == nil {
		ctx = context.WithValue(ctx, tplctx.TemplateListSelfOnly, listSelfOnly)
	}

	templates, err := a.templateCtl.ListTemplate(ctx)
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
	op := "template: get template schema"
	r := c.Param(_releaseParam)
	// get template schema by templateName and releaseName
	params := make(map[string]string)

	var (
		releaseID uint64
		err       error
	)

	if releaseID, err = strconv.ParseUint(r, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("releaseID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("releaseID not found or invalid"))
		return
	}

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

	schema, err := a.templateCtl.GetTemplateSchema(c, uint(releaseID), params)
	if err != nil {
		response.AbortWithError(c, err)
		return
	}
	response.SuccessWithData(c, schema)
}

func (a *API) CreateTemplate(c *gin.Context) {
	op := "template: create template"

	g := c.Param(_groupParam)
	var (
		groupID uint64
		err     error
	)
	if groupID, err = strconv.ParseUint(g, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("clusterID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("clusterID not found or invalid"))
		return
	}

	// TODO(zhixiang): name and tag validation
	var createRequest templatectl.CreateTemplateRequest
	err = c.ShouldBindJSON(&createRequest)
	if err != nil {
		log.WithFiled(c, "op", op).Infof("request body is invalid %s", err)
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("request body is invalid"))
		return
	}

	template, err := a.templateCtl.CreateTemplate(c, uint(groupID), createRequest)
	if err != nil {
		if perror.Cause(err) == herrors.ErrNoPrivilege {
			log.WithFiled(c, "op", op).Info("non-admin user try to access root group")
			response.AbortWithRPCError(c, rpcerror.ForbiddenError.WithErrMsg(fmt.Sprintf("no privilege: %s", err.Error())))
			return
		}
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("group with ID %d not found", groupID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}

	_, err = a.templateCtl.CreateRelease(c, template.ID, createRequest.CreateReleaseRequest)
	if err != nil {
		defer func() { _ = a.templateCtl.DeleteTemplate(c, template.ID) }()

		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Infof("could not parse gitlab url: %s", err)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("failed parsing gitlab URL: %s", err)))
			return
		}
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("template with ID %d not found", template.ID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		if _, ok := perror.Cause(err).(*herrors.HorizonErrCreateFailed); ok {
			log.WithFiled(c, "op", op).Infof("release named %s duplicated", createRequest.Name)
			response.AbortWithRPCError(c, rpcerror.ParamError.
				WithErrMsg(fmt.Sprintf("release named %s was duplicate", createRequest.Name)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.SuccessWithData(c, template)
}

func (a *API) ListTemplatesByGroupID(c *gin.Context) {
	op := "template: list templates by group id"

	g := c.Param(_groupParam)

	withFullPathStr := c.Query(_withFullPath)
	withFullPath, err := strconv.ParseBool(withFullPathStr)
	var ctx context.Context = c
	if err == nil {
		ctx = context.WithValue(ctx, tplctx.TemplateWithFullPath, withFullPath)
	}

	listRecursivelyStr := c.Query(_listRecursively)
	listRecursively, err := strconv.ParseBool(listRecursivelyStr)
	if err == nil {
		ctx = context.WithValue(ctx, tplctx.TemplateListRecursively, listRecursively)
	}

	var groupID uint64
	if groupID, err = strconv.ParseUint(g, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("clusterID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("clusterID not found or invalid"))
		return
	}

	var templates templatectl.Templates
	if templates, err = a.templateCtl.ListTemplateByGroupID(ctx, uint(groupID), false); err != nil {
		if perror.Cause(err) == herrors.ErrNoPrivilege {
			log.WithFiled(c, "op", op).Info("non-admin user try to access root group")
			response.AbortWithRPCError(c, rpcerror.ForbiddenError.WithErrMsg(fmt.Sprintf("no privilege: %s", err.Error())))
			return
		}
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("group with ID %d not found", groupID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.SuccessWithData(c, templates)
}

func (a *API) GetTemplate(c *gin.Context) {
	op := "template: get template"

	t := c.Param(_templateParam)
	var (
		templateID uint64
		err        error
	)

	withReleasesStr := c.Query(_withReleases)
	withReleases, err := strconv.ParseBool(withReleasesStr)
	var ctx context.Context = c
	if err == nil {
		ctx = context.WithValue(ctx, tplctx.TemplateWithRelease, withReleases)
	}

	if templateID, err = strconv.ParseUint(t, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("templateID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("templateID not found or invalid"))
		return
	}

	var template *templatectl.Template
	if template, err = a.templateCtl.GetTemplate(ctx, uint(templateID)); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("template with ID %d not found", templateID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.SuccessWithData(c, template)
}

func (a *API) UpdateTemplate(c *gin.Context) {
	op := "template: update template"

	t := c.Param(_templateParam)
	var (
		templateID uint64
		err        error
	)

	if templateID, err = strconv.ParseUint(t, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("templateID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("templateID not found or invalid"))
		return
	}

	var updateRequest templatectl.UpdateTemplateRequest
	err = c.ShouldBindJSON(&updateRequest)
	if err != nil {
		log.WithFiled(c, "op", op).Infof("request body is invalid %s", err)
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("request body is invalid"))
		return
	}

	if err = a.templateCtl.UpdateTemplate(c, uint(templateID), updateRequest); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("template with ID %d not found", templateID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.Success(c)
}

func (a *API) DeleteTemplate(c *gin.Context) {
	op := "template: delete template"

	t := c.Param(_templateParam)
	var (
		templateID uint64
		err        error
	)

	if templateID, err = strconv.ParseUint(t, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("templateID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("templateID not found or invalid"))
		return
	}

	if err = a.templateCtl.DeleteTemplate(c, uint(templateID)); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("template with ID %d not found", templateID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		} else if e := perror.Cause(err); e == herrors.ErrSubResourceExist {
			response.AbortWithRPCError(c, rpcerror.BadRequestError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.Success(c)
}

func (a *API) CreateRelease(c *gin.Context) {
	op := "template: create release"

	t := c.Param(_templateParam)
	var (
		templateID uint64
		err        error
	)
	if templateID, err = strconv.ParseUint(t, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("clusterID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("clusterID not found or invalid"))
		return
	}

	var createRequest templatectl.CreateReleaseRequest
	err = c.ShouldBindJSON(&createRequest)
	if err != nil {
		log.WithFiled(c, "op", op).Infof("request body is invalid %s", err)
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("request body is invalid"))
		return
	}

	var release *templatectl.Release
	if release, err = a.templateCtl.CreateRelease(c, uint(templateID), createRequest); err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			log.WithFiled(c, "op", op).Infof("could not parse gitlab url: %s", err)
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(fmt.Sprintf("failed parsing gitlab URL: %s", err)))
			return
		}
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("template with ID %d not found", templateID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		if _, ok := perror.Cause(err).(*herrors.HorizonErrCreateFailed); ok {
			log.WithFiled(c, "op", op).Infof("release named %s duplicated", createRequest.Name)
			response.AbortWithRPCError(c, rpcerror.ParamError.
				WithErrMsg(fmt.Sprintf("release named %s was duplicate", createRequest.Name)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.SuccessWithData(c, release)
}

func (a *API) GetReleases(c *gin.Context) {
	op := "template: get releases"

	t := c.Param(_templateParam)
	var (
		templateID uint64
		err        error
		releases   templatectl.Releases
	)
	if templateID, err = strconv.ParseUint(t, 10, 64); err != nil {
		// FIXME
		releases, err = a.templateCtl.ListTemplateRelease(c, t)
		if err != nil {
			log.WithFiled(c, "op", op).Errorf("%+v", err)
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
			return
		}
		response.SuccessWithData(c, releases)
		return
	}

	if releases, err = a.templateCtl.ListTemplateReleaseByTemplateID(c, uint(templateID)); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("template with ID %d not found", templateID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.SuccessWithData(c, releases)
}

func (a *API) DeleteRelease(c *gin.Context) {
	op := "template: delete release"

	r := c.Param(_releaseParam)
	var (
		releaseID uint64
		err       error
	)

	if releaseID, err = strconv.ParseUint(r, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("releaseID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("releaseID not found or invalid"))
		return
	}

	if err = a.templateCtl.DeleteRelease(c, uint(releaseID)); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("release with ID %d not found", releaseID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		} else if e := perror.Cause(err); e == herrors.ErrSubResourceExist {
			response.AbortWithRPCError(c, rpcerror.BadRequestError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.Success(c)
}

func (a *API) GetRelease(c *gin.Context) {
	op := "template: get release"

	r := c.Param(_releaseParam)
	var (
		releaseID uint64
		err       error
	)

	if releaseID, err = strconv.ParseUint(r, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("releaseID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("releaseID not found or invalid"))
		return
	}

	var release *templatectl.Release
	if release, err = a.templateCtl.GetRelease(c, uint(releaseID)); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("release with ID %d not found", releaseID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.SuccessWithData(c, release)
}

func (a *API) UpdateRelease(c *gin.Context) {
	op := "template: update release"

	r := c.Param(_releaseParam)
	var (
		releaseID uint64
		err       error
	)

	if releaseID, err = strconv.ParseUint(r, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("releaseID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("releaseID not found or invalid"))
		return
	}

	var updateRequest templatectl.UpdateReleaseRequest
	err = c.ShouldBindJSON(&updateRequest)
	if err != nil {
		log.WithFiled(c, "op", op).Infof("request body is invalid %s", err)
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("request body is invalid"))
		return
	}

	if err = a.templateCtl.UpdateRelease(c, uint(releaseID), updateRequest); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("release with ID %d not found", releaseID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.Success(c)
}

func (a *API) SyncReleaseToRepo(c *gin.Context) {
	op := "template: sync release to repo"

	r := c.Param(_releaseParam)
	var (
		releaseID uint64
		err       error
	)

	if releaseID, err = strconv.ParseUint(r, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("releaseID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("releaseID not found or invalid"))
		return
	}

	if err = a.templateCtl.SyncReleaseToRepo(c, uint(releaseID)); err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("release with ID %d not found", releaseID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.Success(c)
}
