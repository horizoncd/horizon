package token

import (
	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/controller/oauthcheck"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/auth"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/log"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/util/sets"
)

var RequestInfoFty auth.RequestInfoFactory

func init() {
	RequestInfoFty = auth.RequestInfoFactory{
		APIPrefixes: sets.NewString("apis"),
	}
}

const (
	CheckResult = "Result"
)

func MiddleWare(oauthCtl oauthcheck.Controller, skipMatchers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// 1. get user from token and set user context
		token, err := common.GetToken(c)
		if err != nil {
			log.Warning(c, "Have not got token")
			c.Next()
			return
		}

		// 2. check token valid
		if err := oauthCtl.ValidateToken(c, token); err != nil {
			if perror.Cause(err) == herrors.ErrOAuthAccessTokenExpired {
				response.AbortWithUnauthorized(c, common.CodeExpired, err.Error())
				return
			}
			if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithUnauthorized(c, common.Unauthorized, e.Error())
				return
			}
			response.AbortWithUnauthorized(c, common.InternalError, err.Error())
			return
		}

		// 3. do scope check(get requestInfo, and do check)
		user, err := oauthCtl.LoadAccessTokenUser(c, token)
		if err != nil {
			if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithUnauthorized(c, common.Unauthorized, e.Error())
				return
			}
			response.AbortWithInternalError(c, err.Error())
			return
		}
		common.SetUser(c, user)

		requestInfo, err := RequestInfoFty.NewRequestInfo(c.Request)
		if err != nil {
			response.AbortWithRequestError(c, common.RequestInfoError, err.Error())
			return
		}
		result, reason, err := oauthCtl.CheckScopePermission(c, token, *requestInfo)
		if err != nil {
			if e, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithUnauthorized(c, common.Unauthorized, e.Error())
				return
			}
		}
		if !result {
			log.WithFiled(c, CheckResult, result).Warningf("reason = %s", reason)
			response.AbortWithForbiddenError(c, common.Forbidden, "")
			return
		}
		log.WithFiled(c, CheckResult, result).Infof("reason = %s", reason)
		c.Next()
	}, skipMatchers...)
}
