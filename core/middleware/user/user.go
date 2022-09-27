package user

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

const (
	HTTPHeaderOperator = "Operator"
	NotAuthHeader      = "X-OIDC-Redirect-To"
)

// Middleware check user is exists in db. If not, add user into db.
// Then attach a User object into context.
func Middleware(param *param.Param, store sessions.Store,
	skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		mgr := param.UserManager
		operator := c.Request.Header.Get(HTTPHeaderOperator)
		// TODO(gjq): remove this later
		// 1. get user by operator if operator is not empty
		if operator != "" {
			u, err := mgr.GetUserByEmail(c, operator)
			if err != nil {
				response.Abort(c, http.StatusUnauthorized,
					http.StatusText(http.StatusUnauthorized), err.Error())
				return
			}
			if u == nil {
				response.Abort(c, http.StatusUnauthorized,
					http.StatusText(http.StatusUnauthorized),
					fmt.Sprintf("no user matched operator: %s", operator))
				return
			}

			c.Set(common.UserContextKey(), &userauth.DefaultInfo{
				Name:     u.Name,
				FullName: u.FullName,
				ID:       u.ID,
				Email:    u.Email,
				Admin:    u.Admin,
			})
			c.Next()
			return
		}

		// 2. token auth request ( get user by token)
		if _, err := common.GetToken(c); err == nil {
			c.Next()
			return
		}

		session, err := store.Get(c.Request, common.CookieKeyAuth)
		if err != nil {
			response.Abort(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized),
				fmt.Sprintf("session is not found\n"+
					"session name = %s\n, err = %v", common.CookieKeyAuth, err))
			return
		}

		u := session.Values[common.SessionKeyAuthUser]
		if user, ok := u.(*userauth.DefaultInfo); ok && user != nil {
			// attach user to context
			common.SetUser(c, user)
			c.Next()
			return
		}

		// default status code of response is 200,
		// if status is not 200, that means it has been handled by other middleware,
		// so just omit it.
		if c.Writer.Status() != http.StatusOK ||
			// if not login, call this to login
			// if signed in, call this to link other api
			c.Request.URL.Path == "/apis/core/v1/login/callback" {
			c.Next()
			return
		}

		c.Header(NotAuthHeader, "NotAuth")
		response.AbortWithRPCError(c, rpcerror.Unauthorized.WithErrMsg("please login"))
	}, skippers...)
}
