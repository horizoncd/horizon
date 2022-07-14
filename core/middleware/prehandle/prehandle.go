package prehandle

import (
	"fmt"
	"path"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/pkg/auth"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/util/sets"
	"github.com/gin-gonic/gin"
)

var RequestInfoFty auth.RequestInfoFactory

func init() {
	RequestInfoFty = auth.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

var (
	resourceApplication = "applications"
	resourceCluster     = "clusters"
)

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

		redirect := false
		id := uint(0)

		if _, err := strconv.Atoi(authRecord.Name); err != nil && authRecord.Name != "" {
			if authRecord.Resource == resourceApplication {
				app, err := mgr.ApplicationManager.GetByName(c, authRecord.Name)
				if err == nil {
					redirect = true
					id = app.ID
				}
			} else if authRecord.Resource == resourceCluster {
				cluster, err := mgr.ClusterMgr.GetByName(c, authRecord.Name)
				if err == nil {
					redirect = true
					id = cluster.ID
				}
			}
		}

		if redirect {
			c.Request.URL.Path = "/" + path.Join(requestInfo.APIPrefix, requestInfo.APIGroup, requestInfo.APIVersion,
				requestInfo.Resource, fmt.Sprintf("%d", id), requestInfo.Subresource)
			r.HandleContext(c)
			c.Abort()
			return
		}

		c.Next()
	}, skippers...)
}
