package jobs

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	jobconfig "github.com/horizoncd/horizon/pkg/config/job"
	"github.com/horizoncd/horizon/pkg/log"
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
