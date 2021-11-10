package common

import "strings"

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
