package autofree

import (
	"context"
	"time"

	"g.hz.netease.com/horizon/core/common"
	clusterctl "g.hz.netease.com/horizon/core/controller/cluster"
	environmentctl "g.hz.netease.com/horizon/core/controller/environment"
	prctl "g.hz.netease.com/horizon/core/controller/pipelinerun"
	"g.hz.netease.com/horizon/lib/q"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	"g.hz.netease.com/horizon/pkg/config/autofree"
	"g.hz.netease.com/horizon/pkg/server/middleware/requestid"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/log"
	uuid "github.com/satori/go.uuid"
)

func AutoReleaseExpiredClusterJob(ctx context.Context, jobConfig *autofree.Config, userMgr usermanager.Manager,
	clusterCtr clusterctl.Controller, prCtr prctl.Controller, envCtr environmentctl.Controller) {
	// verify account
	user, err := userMgr.GetUserByIDP(ctx, jobConfig.Account, jobConfig.AccountIDP)
	if err != nil {
		log.Errorf(ctx, "failed to verify operator, err: %v", err.Error())
		panic(err)
	}
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:     user.Name,
		FullName: user.FullName,
		ID:       user.ID,
		Email:    user.Email,
		Admin:    user.Admin,
	})

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

func process(ctx context.Context, jobConfig *autofree.Config, clusterCtr clusterctl.Controller,
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
				supported := func() bool {
					for _, env := range jobConfig.SupportedEnvs {
						if clr.EnvironmentName == env {
							return true
						}
					}
					log.WithFiled(ctx, "op", op).
						Warningf("%v environment does not allow auto-free. cluster: %v, expire seconds: %v",
							clr.EnvironmentName, clr.Name, clr.ExpireSeconds)
					return false
				}()
				environment, err := envCtr.GetByName(ctx, clr.EnvironmentName)
				if !supported || err != nil || !environment.AutoFree {
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
					log.WithFiled(ctx, "op", op).Errorf("failed to automatically release cluster: %v, err: %v", clr.Name, err.Error())
				} else {
					log.WithFiled(ctx, "op", op).Infof("cluster %v automatic releasing succeeded", clr.Name)
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
