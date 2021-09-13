package user

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/common"
	"g.hz.netease.com/horizon/pkg/config/oidc"
	"g.hz.netease.com/horizon/pkg/user"
	"g.hz.netease.com/horizon/pkg/user/models"
	"g.hz.netease.com/horizon/server/middleware"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const contextUserKey = "contextUser"

// Middleware check user is exists in db. If not, add user into db.
// Then attach a User object into context.
func Middleware(config oidc.Config, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		oidcID := c.Request.Header.Get(config.OIDCIDHeader)
		oidcType := c.Request.Header.Get(config.OIDCTypeHeader)
		userName := c.Request.Header.Get(config.UserHeader)
		email := c.Request.Header.Get(config.EmailHeader)

		// if one of the fields is empty, return 401 Unauthorized
		if len(oidcID) == 0 || len(oidcType) == 0 || len(userName) == 0 || len(email) == 0 {
			response.Abort(c, http.StatusUnauthorized,
				http.StatusText(http.StatusUnauthorized), http.StatusText(http.StatusUnauthorized))
			return
		}

		mgr := user.Mgr
		u, err := mgr.GetByOIDCMeta(c, oidcID, oidcType)
		if err != nil {
			response.AbortWithInternalError(c, common.FindUserError, fmt.Sprintf("error to find user: %v", err))
			return
		}
		if u == nil {
			u, err = mgr.Create(c, &models.User{
				Name:     userName,
				Email:    email,
				OIDCId:   oidcID,
				OIDCType: oidcType,
			})
			if err != nil {
				response.AbortWithInternalError(c, common.CreateUserError, fmt.Sprintf("error to create user: %v", err))
				return
			}
		}
		// attach user to context
		c.Set(contextUserKey, u)
		c.Next()
	}, skippers...)
}

func FromContext(ctx context.Context) (*models.User, error) {
	u, ok := ctx.Value(contextUserKey).(*models.User)
	if !ok {
		return nil, errors.New("cannot get the user from context")
	}
	return u, nil
}
