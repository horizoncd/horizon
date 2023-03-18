package main

import (
	"github.com/horizoncd/horizon/job/cmd"

	// for image registry
	_ "github.com/horizoncd/horizon/pkg/cluster/registry/harbor"

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
	cmd.Run()
}
