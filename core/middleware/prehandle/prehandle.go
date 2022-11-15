package prehandle

import (
	"fmt"
	"path"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/auth"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/sets"
	"github.com/gin-gonic/gin"
)

const (
	pathReleaseSchema = "schema"
)

var RequestInfoFty auth.RequestInfoFactory

func init() {
	RequestInfoFty = auth.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

func Middleware(r *gin.Engine, mgr *managerparam.Manager, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		requestInfo, err := RequestInfoFty.NewRequestInfo(c.Request)
		if err != nil {
			response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
			return
		}

		// 2. construct record
		authRecord := auth.AttributesRecord{
			Verb:            requestInfo.Verb,
			APIGroup:        requestInfo.APIGroup,
			APIVersion:      requestInfo.APIVersion,
			Resource:        requestInfo.Resource,
			SubResource:     requestInfo.Subresource,
			Name:            requestInfo.Name,
			Scope:           requestInfo.Scope,
			ResourceRequest: requestInfo.IsResourceRequest,
			Path:            requestInfo.Path,
		}
		c.Set(common.ContextAuthRecord, authRecord)

		if _, err := strconv.Atoi(authRecord.Name); err != nil &&
			authRecord.Name != "" && authRecord.APIGroup == common.GroupCore {
			if authRecord.Resource == common.ResourceApplication {
				handleApplication(c, mgr, r, authRecord, requestInfo)
				return
			} else if authRecord.Resource == common.ResourceCluster {
				handleCluster(c, mgr, r, authRecord, requestInfo)
				return
			} else if authRecord.Resource == common.ResourceTemplate {
				if authRecord.SubResource == common.AliasTemplateRelease &&
					len(requestInfo.Parts) == 5 && requestInfo.Parts[4] == pathReleaseSchema {
					handleGetSchema(c, mgr, r, authRecord, requestInfo)
					return
				}
				handleTemplate(c, mgr, r, authRecord, requestInfo)
				return
			}
		}

		if authRecord.APIGroup == common.GroupFront {
			if authRecord.Resource == common.ResourceCluster {
				handleFrontCluster(c, r)
			} else if authRecord.Resource == common.ResourceApplication {
				handleFrontApplication(c, r)
			}
		}
	}, skippers...)
}

func constructRBACParam(c *gin.Context) (*auth.AttributesRecord, error) {
	requestInfo, err := RequestInfoFty.NewRequestInfo(c.Request)
	if err != nil {
		return nil, err
	}

	// 2. construct record
	authRecord := auth.AttributesRecord{
		Verb:            requestInfo.Verb,
		APIGroup:        requestInfo.APIGroup,
		APIVersion:      requestInfo.APIVersion,
		Resource:        requestInfo.Resource,
		SubResource:     requestInfo.Subresource,
		Name:            requestInfo.Name,
		Scope:           requestInfo.Scope,
		ResourceRequest: requestInfo.IsResourceRequest,
		Path:            requestInfo.Path,
	}
	return &authRecord, nil
}

func handleFrontCluster(c *gin.Context, r *gin.Engine) {
	if c.Request.URL.Path == "/apis/front/v1/clusters/searchmyclusters" {
		currentUser, err := common.UserFromContext(c)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.InternalError.WithErrMsgf(
					"current user not found\n"+
						"err = %v", err))
			return
		}
		c.Request.URL.RawQuery += fmt.Sprintf("&%s=%d", common.ClusterQueryByUser, currentUser.GetID())
		c.Request.URL.Path = "/apis/core/v1/clusters"
		r.HandleContext(c)
		return
	} else if c.Request.URL.Path == "/apis/front/v1/clusters/searchclusters" {
		c.Request.URL.Path = "/apis/core/v1/clusters"
		r.HandleContext(c)
		return
	}
}

func handleFrontApplication(c *gin.Context, r *gin.Engine) {
	if c.Request.URL.Path == "/apis/front/v1/applications/searchmyapplications" {
		currentUser, err := common.UserFromContext(c)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.InternalError.WithErrMsgf(
					"current user not found\n"+
						"err = %v", err))
			return
		}
		c.Request.URL.RawQuery += fmt.Sprintf("&%s=%d", common.ClusterQueryByUser, currentUser.GetID())
		c.Request.URL.Path = "/apis/core/v1/applications"
		r.HandleContext(c)
		return
	} else if c.Request.URL.Path == "/apis/front/v1/applications/searchapplications" {
		c.Request.URL.Path = "/apis/core/v1/applications"
		r.HandleContext(c)
		return
	}
}

func handleApplication(c *gin.Context, mgr *managerparam.Manager, r *gin.Engine,
	authRecord auth.AttributesRecord, requestInfo *auth.RequestInfo) {
	app, err := mgr.ApplicationManager.GetByName(c, authRecord.Name)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ApplicationInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
	} else {
		c.Request.URL.Path = "/" + path.Join(requestInfo.APIPrefix, requestInfo.APIGroup, requestInfo.APIVersion,
			requestInfo.Resource, fmt.Sprintf("%d", app.ID), requestInfo.Subresource)
		for i, param := range c.Params {
			if param.Key == common.ParamApplicationID {
				c.Params[i].Value = fmt.Sprintf("%d", app.ID)
			}
		}
	}
	authRecordPtr, err := constructRBACParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
		return
	}
	c.Set(common.ContextAuthRecord, *authRecordPtr)
	c.Next()
}

func handleCluster(c *gin.Context, mgr *managerparam.Manager, r *gin.Engine,
	authRecord auth.AttributesRecord, requestInfo *auth.RequestInfo) {
	cluster, err := mgr.ClusterMgr.GetByName(c, authRecord.Name)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.ClusterInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
	} else {
		c.Request.URL.Path = "/" + path.Join(requestInfo.APIPrefix, requestInfo.APIGroup, requestInfo.APIVersion,
			requestInfo.Resource, fmt.Sprintf("%d", cluster.ID), requestInfo.Subresource)
		for i, param := range c.Params {
			if param.Key == common.ParamClusterID {
				c.Params[i].Value = fmt.Sprintf("%d", cluster.ID)
			}
		}
	}
	authRecordPtr, err := constructRBACParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
		return
	}
	c.Set(common.ContextAuthRecord, *authRecordPtr)
	c.Next()
}

func handleGetSchema(c *gin.Context, mgr *managerparam.Manager, r *gin.Engine,
	authRecord auth.AttributesRecord, requestInfo *auth.RequestInfo) {
	template, err := mgr.TemplateMgr.GetByName(c, authRecord.Name)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.TemplateInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		c.Next()
		return
	}

	release, err := mgr.TemplateReleaseManager.
		GetByTemplateNameAndRelease(c, authRecord.Name, requestInfo.Parts[3])
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.TemplateReleaseInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		c.Next()
		return
	}

	c.Request.URL.Path = "/" + path.Join(requestInfo.APIPrefix, requestInfo.APIGroup, requestInfo.APIVersion,
		requestInfo.Resource, fmt.Sprintf("%d", template.ID),
		requestInfo.Subresource, fmt.Sprintf("%d", release.ID), pathReleaseSchema)
	for i, param := range c.Params {
		if param.Key == common.ParamTemplateID {
			c.Params[i].Value = fmt.Sprintf("%d", template.ID)
		}
		if param.Key == common.ParamReleaseID {
			c.Params[i].Value = fmt.Sprintf("%d", release.ID)
		}
	}
	authRecordPtr, err := constructRBACParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
		return
	}
	c.Set(common.ContextAuthRecord, *authRecordPtr)
	c.Next()
}

func handleTemplate(c *gin.Context, mgr *managerparam.Manager, r *gin.Engine,
	authRecord auth.AttributesRecord, requestInfo *auth.RequestInfo) {
	template, err := mgr.TemplateMgr.GetByName(c, authRecord.Name)
	if err != nil {
		if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok && e.Source == herrors.TemplateInDB {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
	} else {
		c.Request.URL.Path = "/" + path.Join(requestInfo.APIPrefix, requestInfo.APIGroup, requestInfo.APIVersion,
			requestInfo.Resource, fmt.Sprintf("%d", template.ID), requestInfo.Subresource)
		for i, param := range c.Params {
			if param.Key == common.ParamTemplateID {
				c.Params[i].Value = fmt.Sprintf("%d", template.ID)
			}
		}
	}
	authRecordPtr, err := constructRBACParam(c)
	if err != nil {
		response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
		return
	}
	c.Set(common.ContextAuthRecord, *authRecordPtr)
	c.Next()
}
