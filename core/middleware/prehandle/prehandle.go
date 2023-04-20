// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package prehandle

import (
	"fmt"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	middleware "github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/auth"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/sets"
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
