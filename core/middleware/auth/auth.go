package auth

import (
	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/core/middleware"
	"g.hz.netease.com/horizon/pkg/auth"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/rbac"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

func Middleware(authorizer rbac.Authorizer, skipMatchers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// get user
		currentUser, err := common.UserFromContext(c)
		if err != nil {
			response.AbortWithForbiddenError(c, common.Forbidden, err.Error())
			return
		}

		record, ok := c.Get(common.ContextAuthRecord)
		if !ok {
			response.AbortWithRPCError(c,
				rpcerror.BadRequestError.WithErrMsg("request with no auth record"))
			return
		}
		authRecord := record.(auth.AttributesRecord)
		authRecord.User = currentUser

		// for routes like /apis/core/v1/applications
		if authRecord.Name == "" && authRecord.IsReadOnly() {
			c.Next()
			return
		}

		decision, reason, err := authorizer.Authorize(c, authRecord)
		if err != nil {
			log.Warningf(c, "auth failed with err = %s", err.Error())
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
				response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
				return
			}
			if perror.Cause(err) == herrors.ErrParamInvalid {
				response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
				return
			}
			response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
			return
		}
		if decision == auth.DecisionDeny {
			log.Warningf(c, "denied request with reason = %s", reason)
			response.AbortWithForbiddenError(c, common.Forbidden, reason)
			return
		}
		log.Debugf(c, "passed request with reason = %s", reason)
		c.Next()
	}, skipMatchers...)
}
