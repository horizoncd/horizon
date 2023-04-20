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

package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	jobconfig "github.com/horizoncd/horizon/pkg/config/job"
	"github.com/horizoncd/horizon/pkg/util/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

type Job = func(ctx context.Context)

// Run runs the job in a single instance
func Run(ctx context.Context, jobconfig *jobconfig.Config, jobs ...Job) {
	// get candidate name
	candidateID := fmt.Sprintf("candidate-%s", uuid.New())

	// get jobconfig in cluster
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Errorf(ctx, "Failed to get jobconfig: %v", err)
		return
	}

	// create the clientset
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}

	// create the leader elector
	var elector *leaderelection.LeaderElector
	electionConfig := leaderelection.LeaderElectionConfig{
		Lock: &resourcelock.LeaseLock{
			LeaseMeta: metav1.ObjectMeta{
				Namespace: jobconfig.LockNS,
				Name:      jobconfig.LockName,
			},
			Client: clientset.CoordinationV1(),
			LockConfig: resourcelock.ResourceLockConfig{
				Identity: candidateID,
			},
		},
		LeaseDuration: time.Duration(jobconfig.LeaseDuration) * time.Second,
		RenewDeadline: time.Duration(jobconfig.RenewDeadline) * time.Second,
		RetryPeriod:   time.Duration(jobconfig.RetryPeriod) * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				for _, job := range jobs {
					go job(ctx)
				}
			},
			OnStoppedLeading: func() {
				log.Debugf(ctx, "%s lost leadership", candidateID)
			},
		},
	}

	elector, err = leaderelection.NewLeaderElector(electionConfig)
	if err != nil {
		panic(err)
	}

	// start the leader elector
	for {
		select {
		case <-ctx.Done():
			return
		default:
			elector.Run(ctx)
		}
	}
}
