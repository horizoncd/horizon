package request

import (
	"fmt"
	"strconv"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/q"
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

	filtersStr := c.Query(common.Filters)
	if filtersStr == "" {
		return nil
	}
	filters := strings.Split(filtersStr, common.FilterGap)
	for _, filter := range filters {
		if filter == "" {
			continue
		}
		filterArr := strings.Split(filter, common.FilterSep)
		if len(filterArr) != 2 || filterArr[0] == "" || filterArr[1] == "" {
			continue
		}
		filterKey, filterValue := filterArr[0], filterArr[1]
		if _, ok := common.FilterKeywords[filterKey]; ok {
			res[filterKey] = filterValue
		}
	}
	return res
}
