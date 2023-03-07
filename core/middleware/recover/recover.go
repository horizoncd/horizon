package recover

import (
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(gin.CustomRecovery(func(c *gin.Context, err interface{}) {
		log.Error(c, string(debug.Stack()))
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg("our server got panic"))
	}), skippers...)
}
