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

package scope

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/applicationregion"
	middleware "github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/auth"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/sets"

	"github.com/gin-gonic/gin"

	herrors "github.com/horizoncd/horizon/core/errors"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	perror "github.com/horizoncd/horizon/pkg/errors"
)

var _createClusterURLPattern = regexp.MustCompile(`/apis/core/v[12]/applications/(\d+)/clusters`)

const (
	_method     = http.MethodPost
	_scopeParam = "scope"
)

var RequestInfoFty auth.RequestInfoFactory

func init() {
	RequestInfoFty = auth.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

// Middleware to set region for create cluster API and set scope param for other cluster APIs
func Middleware(param *param.Param, applicationRegionCtl applicationregion.Controller,
	mgr *managerparam.Manager, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// for request to create cluster, set default region to scope if not provided
		if _createClusterURLPattern.MatchString(c.Request.URL.Path) && c.Request.Method == _method {
			matches := _createClusterURLPattern.FindStringSubmatch(c.Request.URL.Path)
			applicationIDStr := matches[1]
			applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
			if err != nil {
				response.AbortWithRequestError(c, common.InvalidRequestParam,
					fmt.Sprintf("invalid applicationID: %v", applicationID))
				return
			}
			// scope format is: {environment}/{region}
			query := c.Request.URL.Query()
			scope := query.Get(_scopeParam)
			params := strings.Split(scope, "/")
			// invalid scope
			if scope == "" || len(params) > 2 {
				response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid scope!")
				return
			}
			// params length is 2, satisfies {environment}/{region} format, skip
			if len(params) == 2 {
				c.Next()
				return
			}
			environment := params[0]

			applicationRegions, err := applicationRegionCtl.List(c, uint(applicationID))
			if err != nil {
				response.AbortWithInternalError(c,
					fmt.Sprintf("failed to get applicationRegions by id: %v", applicationID))
				return
			}
			r, ok := getRegion(c, applicationRegions, environment, param)
			if !ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(
					fmt.Sprintf("cannot find region for environment %v, application %v",
						environment, applicationID)))
				return
			}

			query.Set(_scopeParam, fmt.Sprintf("%v/%v", environment, r))
			c.Request.URL.RawQuery = query.Encode()
			log.Debugf(c, "success to set default region to param: %s", query.Get(_scopeParam))
			c.Next()
			return
		}
		requestInfo, err := RequestInfoFty.NewRequestInfo(c.Request)
		if err != nil {
			response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
			return
		}
		// for request to access cluster and its sub resources, set scope param
		if requestInfo.APIGroup == common.GroupCore && (requestInfo.Resource == common.ResourceCluster ||
			requestInfo.Resource == common.ResourcePipelinerun ||
			requestInfo.Resource == common.ResourceCheckrun) && requestInfo.Name != "" {
			query := c.Request.URL.Query()
			if query.Get(_scopeParam) == "" {
				scope, err := getScope(c, mgr, requestInfo.Resource, requestInfo.Name)
				if err != nil {
					if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
						response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
						return
					}
					log.Errorf(c, "failed to get scope: %v", err.Error())
					c.Next()
					return
				}
				query.Set(_scopeParam, scope)
				c.Request.URL.RawQuery = query.Encode()
				log.Debugf(c, "success to set scope to param: %s", query.Get(_scopeParam))
				c.Next()
			}
		}
		c.Next()
	}, skippers...)
}

func getScope(c *gin.Context, mgr *managerparam.Manager, resourceType, resourceName string) (string, error) {
	var (
		resourceID uint
		cluster    *clustermodels.Cluster
		err        error
	)
	id, err := strconv.Atoi(resourceName)
	if err == nil {
		resourceID = uint(id)
		if resourceType == common.ResourceCheckrun {
			checkrun, err := mgr.PRMgr.Check.GetCheckRunByID(c, resourceID)
			if err != nil {
				return "", err
			}
			resourceID = checkrun.PipelineRunID
			resourceType = common.ResourcePipelinerun
		}
		if resourceType == common.ResourcePipelinerun {
			pipelineRun, err := mgr.PRMgr.PipelineRun.GetByID(c, resourceID)
			if err != nil {
				return "", err
			}
			resourceID = pipelineRun.ClusterID
			resourceType = common.ResourceCluster
		}
		if resourceType == common.ResourceCluster {
			cluster, err = mgr.ClusterMgr.GetByID(c, resourceID)
			if err != nil {
				return "", err
			}
		}
	} else {
		if resourceType != common.ResourceCluster {
			return "", herrors.ErrUnsupportedResourceType
		}
		cluster, err = mgr.ClusterMgr.GetByName(c, resourceName)
		if err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("%s/%s", cluster.EnvironmentName, cluster.RegionName), nil
}

func getRegion(c *gin.Context, applicationRegions applicationregion.ApplicationRegion,
	environment string, p *param.Param) (string, bool) {
	for _, applicationRegion := range applicationRegions {
		if applicationRegion.Environment == environment {
			r := applicationRegion.Region
			region, err := p.RegionMgr.GetRegionByName(c, r)
			if err != nil {
				log.Errorf(c, "query region failed: %s, err: %+v", r, err)
				return "", false
			}
			if region.Disabled {
				log.Errorf(c, "region disabled: %s", r)
				return "", false
			}
			return r, true
		}
	}
	return "", false
}
