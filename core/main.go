package main

import (
	_ "net/http/pprof"

	"g.hz.netease.com/horizon/core/cmd"
)

func main() {
	cmd.Run(cmd.ParseFlags())
}
