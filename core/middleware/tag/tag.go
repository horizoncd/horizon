package tag

import (
	"fmt"

	"g.hz.netease.com/horizon/core/common"
	middleware "g.hz.netease.com/horizon/core/middleware"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/tag/models"
	"g.hz.netease.com/horizon/pkg/util/sets"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/labels"
)

// Middleware to parse tagSelector params
func Middleware(skippers ...middleware.Skipper) gin.HandlerFunc {
	return middleware.New(func(c *gin.Context) {
		// parse tagSelector
		tagSelectorStr := c.Request.URL.Query().Get(common.TagSelector)
		if tagSelectorStr != "" {
			k8sSelector, err := labels.Parse(tagSelectorStr)
			if err != nil {
				response.AbortWithRPCError(c,
					rpcerror.BadRequestError.WithErrMsg(fmt.Sprintf("invalid tagSelector, err: %s", err.Error())))
			}
			requirements, selectable := k8sSelector.Requirements()
			if !selectable || len(requirements) == 0 {
				return
			}

			var tagSelectors []models.TagSelector
			for _, requirement := range requirements {
				values := sets.NewString(requirement.Values().List()...)
				tagSelectors = append(tagSelectors, models.TagSelector{
					Key:      requirement.Key(),
					Values:   values,
					Operator: string(requirement.Operator()),
				})
			}

			c.Set(common.TagSelector, tagSelectors)
		}
		c.Next()
	}, skippers...)
}
