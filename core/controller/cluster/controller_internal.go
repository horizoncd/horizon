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

package cluster

import (
	"context"
	"fmt"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/cd"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/gitrepo"
	eventmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

func (c *controller) InternalDeploy(ctx context.Context, clusterID uint,
	r *InternalDeployRequest) (_ *InternalDeployResponse, err error) {
	const op = "cluster controller: internal deploy"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get pr, and do some validate
	pr, err := c.pipelinerunMgr.GetByID(ctx, r.PipelinerunID)
	if err != nil {
		return nil, err
	}
	if pr == nil || pr.ClusterID != clusterID {
		return nil, herrors.NewErrNotFound(herrors.Pipelinerun,
			fmt.Sprintf("cannot find the pipelinerun with id: %v", r.PipelinerunID))
	}

	// 2. get some relevant models
	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. update image in git repo
	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, cluster.Template, cluster.TemplateRelease)
	if err != nil {
		return nil, err
	}
	po := gitrepo.PipelineOutput{
		Image: &pr.ImageURL,
		Git: &gitrepo.Git{
			URL:      &pr.GitURL,
			CommitID: &pr.GitCommit,
		},
	}
	switch pr.GitRefType {
	case codemodels.GitRefTypeTag:
		po.Git.Tag = &pr.GitRef
	case codemodels.GitRefTypeBranch:
		po.Git.Branch = &pr.GitRef
	}
	commit, err := c.clusterGitRepo.UpdatePipelineOutput(ctx, application.Name, cluster.Name,
		tr.ChartName, po)
	if err != nil {
		return nil, perror.WithMessage(err, op)
	}

	// 4. update config commit and status
	if err := c.pipelinerunMgr.UpdateConfigCommitByID(ctx, pr.ID, commit); err != nil {
		return nil, err
	}
	updatePRStatus := func(pState eventmodels.PipelineStatus, revision string) error {
		if err = c.pipelinerunMgr.UpdateStatusByID(ctx, pr.ID, pState); err != nil {
			log.Errorf(ctx, "UpdateStatusByID error, pr = %d, status = %s, err = %v",
				pr.ID, pState, err)
			return err
		}
		log.Infof(ctx, "InternalDeploy status, pr = %d, status = %s, revision = %s",
			pr.ID, pState, revision)
		return nil
	}
	if err := updatePRStatus(eventmodels.StatusCommitted, commit); err != nil {
		return nil, err
	}

	// 5. merge branch from gitops to master  and update status
	masterRevision, err := c.clusterGitRepo.MergeBranch(ctx, application.Name, cluster.Name,
		gitrepo.GitOpsBranch, c.clusterGitRepo.DefaultBranch(), &pr.ID)
	if err != nil {
		return nil, perror.WithMessage(err, op)
	}
	if err := updatePRStatus(eventmodels.StatusMerged, masterRevision); err != nil {
		return nil, err
	}

	// 6. create cluster in cd system
	regionEntity, err := c.regionMgr.GetRegionEntity(ctx, cluster.RegionName)
	if err != nil {
		return nil, err
	}
	envValue, err := c.clusterGitRepo.GetEnvValue(ctx, application.Name, cluster.Name, tr.ChartName)
	if err != nil {
		return nil, err
	}
	repoInfo := c.clusterGitRepo.GetRepoInfo(ctx, application.Name, cluster.Name)
	if err := c.cd.CreateCluster(ctx, &cd.CreateClusterParams{
		Environment:  cluster.EnvironmentName,
		Cluster:      cluster.Name,
		GitRepoURL:   repoInfo.GitRepoURL,
		ValueFiles:   repoInfo.ValueFiles,
		RegionEntity: regionEntity,
		Namespace:    envValue.Namespace,
	}); err != nil {
		return nil, err
	}

	// 7. reset cluster status
	if cluster.Status == common.ClusterStatusFreed {
		cluster.Status = common.ClusterStatusEmpty
		cluster, err = c.clusterMgr.UpdateByID(ctx, cluster.ID, cluster)
		if err != nil {
			return nil, err
		}
	}

	// 8. deploy cluster in cd system
	if err := c.cd.DeployCluster(ctx, &cd.DeployClusterParams{
		Environment: cluster.EnvironmentName,
		Cluster:     cluster.Name,
		Revision:    masterRevision,
	}); err != nil {
		return nil, err
	}

	// 9. update status
	if err := updatePRStatus(eventmodels.StatusOK, masterRevision); err != nil {
		return nil, err
	}

	// 10. record event
	if _, err := c.eventMgr.CreateEvent(ctx, &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			EventType:    eventmodels.ClusterBuildDeployed,
			ResourceID:   cluster.ID,
		},
	}); err != nil {
		log.Warningf(ctx, "failed to create event, err: %s", err.Error())
	}

	return &InternalDeployResponse{
		PipelinerunID: pr.ID,
		Commit:        commit,
	}, nil
}
