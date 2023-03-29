package request

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/core/common"
)

// GetPageParam get and check the page params
func GetPageParam(c *gin.Context) (int, int, error) {
	var (
		pageNumber int
		pageSize   int
		err        error
	)
	pNumber := c.Query(common.PageNumber)
	if pNumber == "" {
		pageNumber = common.DefaultPageNumber
	} else {
		pageNumber, err = strconv.Atoi(pNumber)
		if err != nil || pageNumber <= 0 {
			return 0, 0, fmt.Errorf("invalid param, pageNumber: %s", pNumber)
		}
	}
	pSize := c.Query(common.PageSize)
	if pSize == "" {
		pageSize = common.DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pSize)
		if err != nil || pageSize <= 0 || pageSize > common.MaxPageSize {
			return 0, 0, fmt.Errorf("invalid param, pageSize: %s", pSize)
		}
	}

	return pageNumber, pageSize, nil
}

func GetFilterParam(c *gin.Context) q.KeyWords {
	var res = make(q.KeyWords)

	template := c.Query(common.Template)
	templateRelease := c.Query(common.TemplateRelease)

	set := func(key, value string) {
		if value == "" {
			return
		}
		res[key] = value
	}
	set(common.Template, template)
	set(common.TemplateRelease, templateRelease)
	return res
}
