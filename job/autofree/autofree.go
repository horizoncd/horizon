package autofree

import (
	"context"
	"time"

	"g.hz.netease.com/horizon/core/common"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	environmentctl "g.hz.netease.com/horizon/core/controller/environment"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	userctl "g.hz.netease.com/horizon/core/controller/user"
	"g.hz.netease.com/horizon/lib/q"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	"g.hz.netease.com/horizon/pkg/util/log"
	uuid "github.com/satori/go.uuid"
)

type Config struct {
	Account       string
	JobInterval   time.Duration
	BatchInterval time.Duration
	BatchSize     int
}

func AutoReleaseExpiredClusterJob(ctx context.Context, jobConfig *Config, userCtr userctl.Controller,
	clusterCtr clusterctl.Controller, prCtr prctl.Controller, envCtr environmentctl.Controller) {
	// verify account
	user, err := userCtr.GetUserByEmail(ctx, jobConfig.Account)
	if err != nil {
		log.Errorf(ctx, "failed to verify operator, err: %v", err.Error())
		panic(err)
	}
	ctx = common.WithContext(ctx, user)

	// start job
	log.Infof(ctx, "Starting releasing expired cluster automatically every %v", jobConfig.JobInterval)
	defer log.Infof(ctx, "Stopping releasing expired cluster automatically")
	ticker := time.NewTicker(jobConfig.JobInterval)
	defer ticker.Stop()
	for range ticker.C {
		rid := uuid.NewV4().String()
		// nolint
		ctx = context.WithValue(ctx, requestid.HeaderXRequestID, rid)
		log.Infof(ctx, "auto-free job starts to execute, rid: %v", rid)
		process(ctx, jobConfig, clusterCtr, prCtr, envCtr)
	}
}

func process(ctx context.Context, jobConfig *Config, clusterCtr clusterctl.Controller,
	prCtr prctl.Controller, envCtr environmentctl.Controller) {
	op := "job: cluster auto-free"
	query := &q.Query{
		PageNumber: common.DefaultPageNumber,
		PageSize:   jobConfig.BatchSize,
		Keywords:   make(map[string]interface{}),
	}
	for {
		// 1. fetch a batch of clusters with expiry
		clusterWithExpiry, err := clusterCtr.ListClusterWithExpiry(ctx, query)
		if err != nil {
			log.WithFiled(ctx, "op", op).
				Errorf("failed to list cluster with expiry, err: %v", err.Error())
			return
		}

		for _, clr := range clusterWithExpiry {
			// 2. Only need to free when the cluster has pipelineruns
			// and expired and its environment supports auto-free
			isNeedFree, err := func() (bool, error) {
				prTotal, pipelineruns, err := prCtr.List(ctx, clr.ID, false, q.Query{
					PageNumber: 1,
					PageSize:   1,
				})
				if err != nil || prTotal == 0 {
					return false, err
				}

				prUpdatedAt := pipelineruns[0].UpdatedAt
				if !expired(clr, prUpdatedAt) {
					return false, nil
				}

				environment, err := envCtr.GetByName(ctx, clr.EnvironmentName)
				if err != nil || !environment.AutoFree || environment.Name == common.OnlineEnv {
					return false, err
				}
				return true, nil
			}()
			if err != nil {
				log.WithFiled(ctx, "op", op).Errorf("%+v", err)
				continue
			}

			// 3. free expired cluster
			if isNeedFree {
				err = clusterCtr.FreeCluster(ctx, clr.ID)
				if err != nil {
					log.WithFiled(ctx, "op", op).Errorf("failed to automaticlly release cluster: %v, err: %v", clr.Name, err.Error())
				} else {
					log.WithFiled(ctx, "op", op).Infof("cluster %v releasing automaticlly succeeded", clr.Name)
				}
			}
		}
		if len(clusterWithExpiry) < query.PageSize {
			break
		}
		query.Keywords[common.IDThan] = clusterWithExpiry[len(clusterWithExpiry)-1].ID
		time.Sleep(jobConfig.BatchInterval)
	}
}

func expired(cluster *clusterctl.ListClusterWithExpiryResponse, prUpdateAt time.Time) bool {
	updatedAt := cluster.UpdatedAt
	var lastUpdateAt time.Time
	if updatedAt.After(prUpdateAt) {
		lastUpdateAt = updatedAt
	} else {
		lastUpdateAt = prUpdateAt
	}
	return lastUpdateAt.Add(time.Duration(cluster.ExpireSeconds * 1e9)).Before(time.Now())
}
