package orm

import (
	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/server/middleware"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Middleware add db to context
func Middleware(db *gorm.DB, skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		_orm := c.Value(orm.ORMKey())
		if _orm == nil {
			c.Set(orm.ORMKey(), db)
		}
		c.Next()
	}, skippers...)
}
