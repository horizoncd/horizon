package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/config/oidc"
	"g.hz.netease.com/horizon/pkg/dao/user"
	middleware2 "g.hz.netease.com/horizon/pkg/server/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"github.com/gin-gonic/gin"
)

const contextUserKey = "contextUser"

// Middleware check user is exists in db. If not, add user into db.
// Then attach a User object into context.
func Middleware(config oidc.Config, skippers ...middleware2.Skipper) gin.HandlerFunc {
	return middleware2.New(func(c *gin.Context) {
		oidcID := c.Request.Header.Get(config.OIDCIDHeader)
		oidcType := c.Request.Header.Get(config.OIDCTypeHeader)
		userName := c.Request.Header.Get(config.UserHeader)
		fullName := c.Request.Header.Get(config.FullNameHeader)
		email := c.Request.Header.Get(config.EmailHeader)

		// if one of the fields is empty, return 401 Unauthorized
		if len(oidcID) == 0 || len(oidcType) == 0 ||
			len(userName) == 0 || len(email) == 0 || len(fullName) == 0 {
			response.Abort(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized), http.StatusText(http.StatusUnauthorized))
			return
		}

		mgr := user.Mgr
		u, err := mgr.GetByOIDCMeta(c, oidcID, oidcType)
		if err != nil {
			response.AbortWithInternalError(c, fmt.Sprintf("error to find user: %v", err))
			return
		}
		if u == nil {
			u, err = mgr.Create(c, &user.User{
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
		c.Set(contextUserKey, &userauth.DefaultInfo{
			Name:     u.Name,
			FullName: u.FullName,
			ID:       int(u.ID),
		})
		c.Next()
	}, skippers...)
}

func FromContext(ctx context.Context) (userauth.User, error) {
	u, ok := ctx.Value(contextUserKey).(userauth.User)
	if !ok {
		return nil, errors.New("cannot get the authenticated user from context")
	}
	return u, nil
}

func Key() string {
	return contextUserKey
}
