package tag

import (
	"fmt"

	"github.com/horizoncd/horizon/core/common"
	middleware "github.com/horizoncd/horizon/core/middleware"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/tag/models"
	"github.com/horizoncd/horizon/pkg/util/sets"

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
