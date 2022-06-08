package orm

import (
	"context"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/server/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Middleware add db to context
func Middleware(db *gorm.DB, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		_orm := c.Value(orm.Key())
		if _orm == nil {
			c.Set(orm.Key(), db)
		}
		c.Next()
	}, skippers...)
}

// MiddlewareSetUserContext set context to db
func MiddlewareSetUserContext(db *gorm.DB, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		_user := c.Value(common.UserContextKey())
		if _user != nil {
			// nolint
			c.Set(orm.Key(), db.WithContext(context.WithValue(context.Background(), common.UserContextKey(), _user)))
		}
		c.Next()
	}, skippers...)
}
