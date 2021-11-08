package request

import (
	"fmt"
	"strconv"

	"g.hz.netease.com/horizon/core/common"
	"github.com/gin-gonic/gin"
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
			return 0, 0, fmt.Errorf("invalid param, pageNumber: %d", pageNumber)
		}
	}
	pSize := c.Query(common.PageSize)
	if pSize == "" {
		pageSize = common.DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pSize)
		if err != nil || pageSize <= 0 || pageSize > common.MaxPageSize {
			return 0, 0, fmt.Errorf("invalid param, pageSize: %d", pageSize)
		}
	}

	return pageNumber, pageSize, nil
}
