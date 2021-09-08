package user

import (
	"fmt"

	"g.hz.netease.com/horizon/pkg/user/dao"
	"g.hz.netease.com/horizon/pkg/user/models"
	"g.hz.netease.com/horizon/server/middleware"
	"g.hz.netease.com/horizon/server/response"
	"github.com/gin-gonic/gin"
)

const (
	HeaderUserName     = "X-HORIZON-OIDC-USER"
	HeaderUserEmail    = "X-HORIZON-OIDC-EMAIL"
	HeaderUserOIDCID   = "X-HORIZON-OIDC-ID"
	HeaderUserOIDCType = "X-HORIZON-OIDC-TYPE"
)

// Middleware check user is exists in db. If not, add user into db.
// Attach a User object into context.
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		oidcID := c.Request.Header.Get(HeaderUserOIDCID)
		oidcType := c.Request.Header.Get(HeaderUserOIDCType)
		userName := c.Request.Header.Get(HeaderUserName)
		email := c.Request.Header.Get(HeaderUserEmail)
		d := dao.New()
		user, err := d.FindByOIDC(c, oidcID, oidcType)
		if err != nil {
			response.AbortWithInternalError(c, "UserFindError", fmt.Sprintf("error to find user: %v", err))
			return
		}
		if user == nil {
			_, err := d.Create(c, &models.User{
				Name:     userName,
				Email:    email,
				OIDCId:   oidcID,
				OIDCType: oidcType,
			})
			if err != nil {
				response.AbortWithInternalError(c, "UserCreateError", fmt.Sprintf("error to create user: %v", err))
				return
			}
		}
		c.Next()
	}, skippers...)
}
