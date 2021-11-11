package cluster

import (
	"context"
	"net/http"
	"time"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	"g.hz.netease.com/horizon/pkg/cluster/cd"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

func (c *controller) Restart(ctx context.Context, clusterID uint) (_ *PipelinerunIDResponse, err error) {
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
	timeNow := time.Now()
	pr := &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionRestart,
		Status:           prmodels.ResultOK,
		Title:            "restart",
		LastConfigCommit: lastConfigCommit.Master,
		ConfigCommit:     commit,
		StartedAt:        &timeNow,
		FinishedAt:       &timeNow,
		CreatedBy:        currentUser.GetID(),
	}
	prCreated, err := c.pipelinerunMgr.Create(ctx, pr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}

func (c *controller) Deploy(ctx context.Context, clusterID uint,
	r *DeployRequest) (_ *PipelinerunIDResponse, err error) {
	const op = "cluster controller: deploy "
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

	// 1. get config commit
	configCommit, err := c.clusterGitRepo.GetConfigCommit(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}
	diff, err := c.clusterGitRepo.CompareConfig(ctx, application.Name, cluster.Name,
		&configCommit.Master, &configCommit.Gitops)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if diff == "" {
		return nil, errors.E(op, http.StatusBadRequest, errors.ErrorCode("NoChange"), "there is no change to deploy")
	}

	// 2. merge branch
	commit, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name)
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

	timeNow := time.Now()
	// 4. add pipelinerun in db
	pr := &prmodels.Pipelinerun{
		ClusterID:        clusterID,
		Action:           prmodels.ActionDeploy,
		Status:           prmodels.ResultOK,
		Title:            r.Title,
		Description:      r.Description,
		LastConfigCommit: configCommit.Master,
		ConfigCommit:     configCommit.Gitops,
		StartedAt:        &timeNow,
		FinishedAt:       &timeNow,
		CreatedBy:        currentUser.GetID(),
	}
	prCreated, err := c.pipelinerunMgr.Create(ctx, pr)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return &PipelinerunIDResponse{
		PipelinerunID: prCreated.ID,
	}, nil
}
