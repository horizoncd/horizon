package cluster

import (
	"context"
	"fmt"

	"g.hz.netease.com/horizon/pkg/cluster/cd"
	"g.hz.netease.com/horizon/pkg/cluster/common"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

func (c *controller) InternalDeploy(ctx context.Context, clusterID uint,
	r *InternalDeployRequest) (_ *InternalDeployResponse, err error) {
	const op = "cluster controller: internal deploy"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get pr, and do some validate
	pr, err := c.pipelinerunMgr.GetByID(ctx, r.PipelinerunID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if pr == nil || pr.ClusterID != clusterID {
		return nil, errors.E(op, fmt.Errorf("cannot find the pipelinerun with id: %v", r.PipelinerunID))
	}

	// 2. get some relevant models
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	er, err := c.envMgr.GetEnvironmentRegionByID(ctx, cluster.EnvironmentRegionID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 3. update image in git repo, and update newest commit to pr
	commit, err := c.clusterGitRepo.UpdatePipelineOutput(ctx, application.Name, cluster.Name, cluster.Template,
		gitrepo.PipelineOutput{
			Image: &pr.ImageURL,
			Git: &gitrepo.Git{
				URL:      &pr.GitURL,
				Branch:   &pr.GitBranch,
				CommitID: &pr.GitCommit,
			},
		})
	if err != nil {
		return nil, errors.E(op, err)
	}
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, pr.ID, commit); err != nil {
		return nil, errors.E(op, err)
	}

	// 4. merge branch from gitops to master
	masterRevision, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// 5. create cluster in cd system
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, er.RegionName)
	if err != nil {
		return nil, errors.E(op, err)
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, cluster.Template)
	if err != nil {
		return nil, errors.E(op, err)
	}
	repoInfo := c.clusterGitRepo.GetRepoInfo(ctx, application.Name, cluster.Name)
	if err := c.cd.CreateCluster(ctx, &cd.CreateClusterParams{
		Environment:   er.EnvironmentName,
		Cluster:       cluster.Name,
		GitRepoSSHURL: repoInfo.GitRepoSSHURL,
		ValueFiles:    repoInfo.ValueFiles,
		RegionEntity:  regionEntity,
		Namespace:     envValue.Namespace,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	// 6. reset cluster status
	if cluster.Status == common.StatusFreed {
		cluster.Status = common.StatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, errors.E(op, err)
		}
	}

	// 7. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: er.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    masterRevision,
	}); err != nil {
		return nil, errors.E(op, err)
	}

	return &InternalDeployResponse{
		PipelinerunID: pr.ID,
		Commit:        commit,
	}, nil
}
