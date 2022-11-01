package main

import (
	_ "net/http/pprof"

	"g.hz.netease.com/horizon/core/cmd"
	_ "g.hz.netease.com/horizon/pkg/cluster/registry/harbor"
)

func main() {
	cmd.Run(cmd.ParseFlags())
}
