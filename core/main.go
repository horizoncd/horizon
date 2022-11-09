package main

import (
	_ "net/http/pprof"

	"g.hz.netease.com/horizon/core/cmd"

	// for image registry
	_ "g.hz.netease.com/horizon/pkg/cluster/registry/harbor"

	// for template repo
	_ "g.hz.netease.com/horizon/pkg/templaterepo/harbor"
)

func main() {
	cmd.Run(cmd.ParseFlags())
}
