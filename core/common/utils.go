package common

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	InternalGitSSHPrefix  string = "ssh://git@g.hz.netease.com:22222"
	InternalGitHTTPPrefix string = "https://g.hz.netease.com"
	CommitHistoryMiddle   string = "/-/commits/"
)

func InternalSSHToHTTPURL(sshURL string) string {
	tmp := strings.TrimPrefix(sshURL, InternalGitSSHPrefix)
	middle := strings.TrimRight(tmp, ".git")
	httpURL := InternalGitHTTPPrefix + middle
	return httpURL
}

// CheckPageParams check whether the params for page is valid
func CheckPageParams(c *gin.Context) (int, int, error) {
	pNumber := c.Query(PageNumber)
	pageNumber, err := strconv.Atoi(pNumber)
	if err != nil || pageNumber <= 0 {
		return 0, 0, fmt.Errorf("invalid param, pageNumber: %d", pageNumber)
	}
	pSize := c.Query(PageSize)
	pageSize, err := strconv.Atoi(pSize)
	if err != nil || pageSize <= 0 || pageSize > MaxPageSize {
		return 0, 0, fmt.Errorf("invalid param, pageSize: %d", pageSize)
	}

	return pageNumber, pageSize, nil
}
