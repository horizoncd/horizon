package cluster

import (
	"context"

	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	ActionBuildDeploy = "builddeploy"
)

func (c *controller) BuildDeploy(ctx context.Context, clusterID uint,
	r *BuildDeployRequest) (_ *BuildDeployResponse, err error) {
	const op = "cluster controller: build deploy"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// action := ActionBuildDeploy

	// cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	// if err != nil {
	// 	return nil, errors.E(op, err)
	// }

	// var branch = cluster.GitBranch
	// if r.Git != nil || r.Git.Branch != "" {
	// 	branch = r.Git.Branch
	// }

	// commit, err := c.commitGetter.GetCommit(ctx, cluster.GitURL, branch)
	// if err != nil {
	// 	return nil, errors.E(op, err)
	// }

	// var lastConfigCommit = ""
	lastPr, err := c.prMgr.GetLastPipelinerunWithConfigCommit(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if lastPr != nil {
		// lastConfigCommit = lastPr.ConfigCommit
	}

	return nil, nil
}
