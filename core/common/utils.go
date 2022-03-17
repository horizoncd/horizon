package common

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"g.hz.netease.com/horizon/pkg/hook/hook"
	"g.hz.netease.com/horizon/pkg/util/log"
)

const (
	InternalGitSSHPrefix  string = "ssh://git@g.hz.netease.com:22222"
	InternalGitHTTPPrefix string = "https://g.hz.netease.com"
	CommitHistoryMiddle   string = "/-/commits/"
)

func InternalSSHToHTTPURL(sshURL string) string {
	tmp := strings.TrimPrefix(sshURL, InternalGitSSHPrefix)
	middle := strings.TrimSuffix(tmp, ".git")
	httpURL := InternalGitHTTPPrefix + middle
	return httpURL
}

func ElegantExit(h hook.Hook) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)

	go func() {
		<-signals
		h.WaitStop()
		os.Exit(0)
	}()
}

func Response(ctx context.Context, resp *http.Response) string {
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Error(ctx, err)
		return err.Error()
	}

	str := string(data)
	log.Info(ctx, str)
	return str
}
