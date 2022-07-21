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
			} else if authRecord.Resource == common.ResourceCluster {
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
			}
		}

		c.Next()
	}, skippers...)
}
