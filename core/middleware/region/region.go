package region

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/applicationregion"
	middleware "github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

// http params for create cluster api
var _urlPattern = regexp.MustCompile(`/apis/core/v[12]/applications/(\d+)/clusters`)

const (
	_method     = http.MethodPost
	_scopeParam = "scope"
)

// Middleware to set region for create cluster API
func Middleware(param *param.Param, applicationRegionCtl applicationregion.Controller,
	skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// not create cluster api, skip
		if !_urlPattern.MatchString(c.Request.URL.Path) || c.Request.Method != _method {
			c.Next()
			return
		}
		matches := _urlPattern.FindStringSubmatch(c.Request.URL.Path)
		applicationIDStr := matches[1]
		applicationID, err := strconv.ParseUint(applicationIDStr, 10, 0)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam,
				fmt.Sprintf("invalid applicationID: %v", applicationID))
			return
		}
		// scope format is: {environment}/{region}
		scope := c.Request.URL.Query().Get(_scopeParam)
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

		c.Request.URL.RawQuery = fmt.Sprintf("%v=%v/%v", _scopeParam, environment, r)
		c.Next()
	}, skippers...)
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
