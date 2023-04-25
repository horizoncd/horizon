// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	_ "github.com/horizoncd/horizon/pkg/workload/deployment"
	_ "github.com/horizoncd/horizon/pkg/workload/kservice"
	_ "github.com/horizoncd/horizon/pkg/workload/pod"
	_ "github.com/horizoncd/horizon/pkg/workload/rollout"
)

func main() {
	cmd.Run(cmd.ParseFlags())
}
