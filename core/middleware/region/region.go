package region

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	appregionmanager "g.hz.netease.com/horizon/pkg/applicationregion/manager"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"
	"g.hz.netease.com/horizon/pkg/config/region"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"

	"github.com/gin-gonic/gin"
)

// http params for create cluster api
var _urlPattern = regexp.MustCompile(`/apis/core/v1/applications/(\d+)/clusters`)

const (
	_method     = http.MethodPost
	_scopeParam = "scope"
)

// Middleware to set region for create cluster API
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
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

		var (
			applicationRegionMgr = appregionmanager.Mgr
		)

		applicationRegions, err := applicationRegionMgr.ListByApplicationID(c, uint(applicationID))
		if err != nil {
			response.AbortWithInternalError(c,
				fmt.Sprintf("failed to get application by id: %v", applicationID))
			return
		}

		config, err := region.FormatEnvironmentDefaultRegions(c)
		if err != nil {
			response.AbortWithInternalError(c, err.Error())
			return
		}
		r := getRegion(applicationRegions, config, environment)
		if len(r) == 0 {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(
				fmt.Sprintf("cannot find region for environment %v, application %v",
					environment, applicationID)))
			return
		}

		c.Request.URL.RawQuery = fmt.Sprintf("%v=%v/%v", _scopeParam, environment, r)
		c.Next()
	}, skippers...)
}

func getRegion(applicationRegions []*models.ApplicationRegion, config *region.Config,
	environment string) string {
	for _, applicationRegion := range applicationRegions {
		if applicationRegion.EnvironmentName == environment {
			return applicationRegion.RegionName
		}
	}
	return getRegionFromConfig(config, environment)
}

func getRegionFromConfig(config *region.Config, environment string) string {
	if config == nil {
		return ""
	}
	// getDefaultRegion get default region of environment
	getDefaultRegion := func(environment string) string {
		return strings.TrimSpace(config.DefaultRegions[environment])
	}

	return getDefaultRegion(environment)
}
