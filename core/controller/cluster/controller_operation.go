package cluster

import (
	"context"
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	timeutil "g.hz.netease.com/horizon/pkg/util/time"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

func (c *controller) Restart(ctx context.Context, clusterID uint) (_ *RestartResponse, err error) {
	const op = "cluster controller: restart "
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 1. get config commit now
	lastConfigCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 2. update restartTime in git repo, and return the newest commit
	commit, err := c.clusterGitRepo.UpdateRestartTime(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    commit,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	// 4. add pipelinerun in db
	pr := &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionRestart,
		Status:           prmodels.ResultOK,
		Title:            fmt.Sprintf("%v-%v", "Restart", timeutil.Now(nil)),
		LastConfigCommit: lastConfigCommit.Master,
		ConfigCommit:     commit,
		CreatedBy:        currentUser.GetID(),
	}
	prCreated, err := c.pipelinerunMgr.Create(ctx, pr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &RestartResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}
