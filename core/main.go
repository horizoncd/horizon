package main

import (
	_ "net/http/pprof"

	"github.com/horizoncd/horizon/core/cmd"

	// for image registry
	_ "github.com/horizoncd/horizon/pkg/cluster/registry/harbor"

	// for template repo
	_ "github.com/horizoncd/horizon/pkg/templaterepo/chartmuseumbase"
)

func main() {
	cmd.Run(cmd.ParseFlags())
}
