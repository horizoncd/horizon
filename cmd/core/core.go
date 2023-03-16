package main

import (
	_ "net/http/pprof"

	"github.com/horizoncd/horizon/core/cmd"

	// for image registry
	_ "github.com/horizoncd/horizon/pkg/cluster/registry/harbor/v1"
	_ "github.com/horizoncd/horizon/pkg/cluster/registry/harbor/v2"

	_ "github.com/horizoncd/horizon/pkg/git"
	_ "github.com/horizoncd/horizon/pkg/git/github"
	_ "github.com/horizoncd/horizon/pkg/git/gitlab"

	// for template repo
	_ "github.com/horizoncd/horizon/pkg/templaterepo/chartmuseumbase"

	// for k8s workload
	_ "github.com/horizoncd/horizon/pkg/cluster/cd/workload/deployment"
	_ "github.com/horizoncd/horizon/pkg/cluster/cd/workload/kservice"
	_ "github.com/horizoncd/horizon/pkg/cluster/cd/workload/pod"
	_ "github.com/horizoncd/horizon/pkg/cluster/cd/workload/rollout"
)

func main() {
	cmd.Run(cmd.ParseFlags())
}
