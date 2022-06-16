package user

import (
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/config/oidc"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/user/models"

	"github.com/gin-gonic/gin"
)

const (
	HTTPHeaderOperator     = "Operator"
	AuthorizationHeaderKey = "Authorization"
)

// Middleware check user is exists in db. If not, add user into db.
// Then attach a User object into context.
func Middleware(param *param.Param, config oidc.Config, skippers ...middleware.Skipper) gin.HandlerFunc {
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

		// 3. else, get by oidc
		oidcID := c.Request.Header.Get(config.OIDCIDHeader)
		oidcType := c.Request.Header.Get(config.OIDCTypeHeader)
		userName := c.Request.Header.Get(config.UserHeader)
		fullName := c.Request.Header.Get(config.FullNameHeader)
		email := c.Request.Header.Get(config.EmailHeader)

		// if one of the fields is empty, return 401 Unauthorized
		// the oidcID will be empty for the common account, such as grp.cloudnative
		if len(oidcType) == 0 || len(userName) == 0 ||
			len(email) == 0 || len(fullName) == 0 {
			response.Abort(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized), http.StatusText(http.StatusUnauthorized))
			return
		}

		u, err := mgr.GetByOIDCMeta(c, oidcType, email)
		if err != nil {
			response.AbortWithInternalError(c, fmt.Sprintf("error to find user: %v", err))
			return
		}
		if u == nil {
			u, err = mgr.Create(c, &models.User{
				Name:     userName,
				FullName: fullName,
				Email:    email,
				OIDCId:   oidcID,
				OIDCType: oidcType,
			})
			if err != nil {
				response.AbortWithInternalError(c, fmt.Sprintf("error to create user: %v", err))
				return
			}
		}
		// attach user to context
		common.SetUser(c, &userauth.DefaultInfo{
			Name:     u.Name,
			FullName: u.FullName,
			ID:       u.ID,
			Email:    u.Email,
			Admin:    u.Admin,
		})
		c.Next()
	}, skippers...)
}
