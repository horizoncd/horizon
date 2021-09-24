package middleware

import (
	"github.com/gin-gonic/gin"
)

// New make a middleware
func New(handler gin.HandlerFunc, skippers ...Skipper) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, skipper := range skippers {
			if skipper(c.Request) {
				c.Next()
				return
			}
		}
		handler(c)
	}
}
