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

package pipelinerun

import (
	"context"
	"fmt"
	"net/http"

	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/cluster/code"
	codemodels "github.com/horizoncd/horizon/pkg/cluster/code"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/collector"
	"github.com/horizoncd/horizon/pkg/cluster/tekton/factory"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/gitrepo"
	appmanager "github.com/horizoncd/horizon/pkg/manager"
	prmodels "github.com/horizoncd/horizon/pkg/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/util/errors"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

type Controller interface {
	GetPipelinerunLog(ctx context.Context, pipelinerunID uint) (*collector.Log, error)
	GetClusterLatestLog(ctx context.Context, clusterID uint) (*collector.Log, error)
	GetDiff(ctx context.Context, pipelinerunID uint) (*GetDiffResponse, error)
	Get(ctx context.Context, pipelinerunID uint) (*PipelineBasic, error)
	List(ctx context.Context, clusterID uint, canRollback bool, query q.Query) (int, []*PipelineBasic, error)
	StopPipelinerun(ctx context.Context, pipelinerunID uint) error
	StopPipelinerunForCluster(ctx context.Context, clusterID uint) error
}

type controller struct {
	pipelinerunMgr appmanager.PipelineRunManager
	applicationMgr appmanager.ApplicationManager
	clusterMgr     appmanager.ClusterManager
	envMgr         appmanager.EnvironmentManager
	tektonFty      factory.Factory
	commitGetter   code.GitGetter
	clusterGitRepo gitrepo.ClusterGitRepo
	userManager    appmanager.UserManager
}

var _ Controller = (*controller)(nil)

func NewController(param *param.Param) Controller {
	return &controller{
		pipelinerunMgr: param.PipelinerunMgr,
		clusterMgr:     param.ClusterMgr,
		envMgr:         param.EnvMgr,
		tektonFty:      param.TektonFty,
		commitGetter:   param.GitGetter,
		applicationMgr: param.ApplicationManager,
		clusterGitRepo: param.ClusterGitRepo,
		userManager:    param.UserManager,
	}
}

func (c *controller) GetPipelinerunLog(ctx context.Context, pipelinerunID uint) (_ *collector.Log, err error) {
	const op = "pipelinerun controller: get pipelinerun log"
	defer wlog.Start(ctx, op).StopPrint()

	pr, err := c.pipelinerunMgr.GetByID(ctx, pipelinerunID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	cluster, err := c.clusterMgr.GetByID(ctx, pr.ClusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}

	// only builddeploy and deploy have logs
	if pr.Action != prmodels.ActionBuildDeploy && pr.Action != prmodels.ActionDeploy {
		return nil, errors.E(op, fmt.Errorf("%v action has no log", pr.Action))
	}

	return c.getPipelinerunLog(ctx, pr, cluster.EnvironmentName)
}

func (c *controller) GetClusterLatestLog(ctx context.Context, clusterID uint) (_ *collector.Log, err error) {
	const op = "pipelinerun controller: get cluster latest log"
	defer wlog.Start(ctx, op).StopPrint()

	pr, err := c.pipelinerunMgr.GetLatestByClusterIDAndActions(ctx, clusterID,
		prmodels.ActionBuildDeploy, prmodels.ActionDeploy)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if pr == nil {
		return nil, errors.E(op, fmt.Errorf("no pipelinerun with log"))
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return c.getPipelinerunLog(ctx, pr, cluster.EnvironmentName)
}

func (c *controller) getPipelinerunLog(ctx context.Context, pr *prmodels.Pipelinerun,
	environment string) (_ *collector.Log, err error) {
	const op = "pipeline controller: get pipelinerun log"
	defer wlog.Start(ctx, op).StopPrint()

	tektonCollector, err := c.tektonFty.GetTektonCollector(environment)
	if err != nil {
		return nil, perror.WithMessagef(err, "failed to get tekton collector for %s", environment)
	}

	return tektonCollector.GetPipelineRunLog(ctx, pr)
}

func (c *controller) GetDiff(ctx context.Context, pipelinerunID uint) (_ *GetDiffResponse, err error) {
	const op = "pipelinerun controller: get pipelinerun diff"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get pipeline
	pipelinerun, err := c.pipelinerunMgr.GetByID(ctx, pipelinerunID)
	if err != nil {
		return nil, err
	}

	// 2. get cluster and application
	cluster, err := c.clusterMgr.GetByID(ctx, pipelinerun.ClusterID)
	if err != nil {
		return nil, err
	}
	application, err := c.applicationMgr.GetByID(ctx, cluster.ApplicationID)
	if err != nil {
		return nil, err
	}

	// 3. get code diff
	var codeDiff *CodeInfo
	if pipelinerun.GitURL != "" && pipelinerun.GitCommit != "" &&
		pipelinerun.GitRef != "" {
		commit, err := c.commitGetter.GetCommit(ctx, pipelinerun.GitURL,
			codemodels.GitRefTypeCommit, pipelinerun.GitCommit)
		if err != nil {
			return nil, err
		}
		historyLink, err := c.commitGetter.GetCommitHistoryLink(pipelinerun.GitURL, pipelinerun.GitCommit)
		if err != nil {
			return nil, err
		}
		codeDiff = &CodeInfo{
			CommitID:  pipelinerun.GitCommit,
			CommitMsg: commit.Message,
			Link:      historyLink,
		}
		switch pipelinerun.GitRefType {
		case codemodels.GitRefTypeTag:
			codeDiff.Tag = pipelinerun.GitRef
		case codemodels.GitRefTypeBranch:
			codeDiff.Branch = pipelinerun.GitRef
		}
	}

	// 4. get config diff
	var configDiff *ConfigDiff
	if pipelinerun.ConfigCommit != "" && pipelinerun.LastConfigCommit != "" {
		diff, err := c.clusterGitRepo.CompareConfig(ctx, application.Name, cluster.Name,
			&pipelinerun.LastConfigCommit, &pipelinerun.ConfigCommit)
		if err != nil {
			return nil, err
		}
		configDiff = &ConfigDiff{
			From: pipelinerun.LastConfigCommit,
			To:   pipelinerun.ConfigCommit,
			Diff: diff,
		}
	}

	return &GetDiffResponse{
		CodeInfo:   codeDiff,
		ConfigDiff: configDiff,
	}, nil
}

func (c *controller) Get(ctx context.Context, pipelineID uint) (_ *PipelineBasic, err error) {
	const op = "pipelinerun controller: get pipelinerun basic"
	defer wlog.Start(ctx, op).StopPrint()

	pipelinerun, err := c.pipelinerunMgr.GetByID(ctx, pipelineID)
	if err != nil {
		return nil, err
	}
	firstCanRollbackPipelinerun, err := c.pipelinerunMgr.GetFirstCanRollbackPipelinerun(ctx, pipelinerun.ClusterID)
	if err != nil {
		return nil, err
	}

	return c.ofPipelineBasic(ctx, pipelinerun, firstCanRollbackPipelinerun)
}

func (c *controller) List(ctx context.Context,
	clusterID uint, canRollback bool, query q.Query) (_ int, _ []*PipelineBasic, err error) {
	const op = "pipelinerun controller: list pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	totalCount, pipelineruns, err := c.pipelinerunMgr.GetByClusterID(ctx, clusterID, canRollback, query)
	if err != nil {
		return 0, nil, err
	}

	// remove the first pipelinerun than can be rollback
	firstCanRollbackPipelinerun, err := c.pipelinerunMgr.GetFirstCanRollbackPipelinerun(ctx, clusterID)
	if err != nil {
		return 0, nil, err
	}

	pipelineBasics, err := c.ofPipelineBasics(ctx, pipelineruns, firstCanRollbackPipelinerun)
	if err != nil {
		return 0, nil, err
	}
	return totalCount, pipelineBasics, nil
}

func (c *controller) ofPipelineBasic(ctx context.Context,
	pr, firstCanRollbackPipelinerun *prmodels.Pipelinerun) (*PipelineBasic, error) {
	user, err := c.userManager.GetUserByID(ctx, pr.CreatedBy)
	if err != nil {
		return nil, err
	}

	canRollback := func() bool {
		// set the firstCanRollbackPipelinerun that cannot rollback
		if firstCanRollbackPipelinerun != nil && pr.ID == firstCanRollbackPipelinerun.ID {
			return false
		}
		return pr.Action != prmodels.ActionRestart && pr.Status == string(prmodels.StatusOK)
	}()

	prBasic := &PipelineBasic{
		ID:               pr.ID,
		Title:            pr.Title,
		Description:      pr.Description,
		Action:           pr.Action,
		Status:           pr.Status,
		GitURL:           pr.GitURL,
		GitCommit:        pr.GitCommit,
		ImageURL:         pr.ImageURL,
		LastConfigCommit: pr.LastConfigCommit,
		ConfigCommit:     pr.ConfigCommit,
		CreatedAt:        pr.CreatedAt,
		UpdatedAt:        pr.UpdatedAt,
		StartedAt:        pr.StartedAt,
		FinishedAt:       pr.FinishedAt,
		CanRollback:      canRollback,
		CreatedBy: UserInfo{
			UserID:   pr.CreatedBy,
			UserName: user.Name,
		},
	}
	switch pr.GitRefType {
	case codemodels.GitRefTypeTag:
		prBasic.GitTag = pr.GitRef
	case codemodels.GitRefTypeBranch:
		prBasic.GitBranch = pr.GitRef
	}
	return prBasic, nil
}

func (c *controller) ofPipelineBasics(ctx context.Context, prs []*prmodels.Pipelinerun,
	firstCanRollbackPipelinerun *prmodels.Pipelinerun) ([]*PipelineBasic, error) {
	var pipelineBasics []*PipelineBasic
	for _, pr := range prs {
		pipelineBasic, err := c.ofPipelineBasic(ctx, pr, firstCanRollbackPipelinerun)
		if err != nil {
			return nil, err
		}
		pipelineBasics = append(pipelineBasics, pipelineBasic)
	}
	return pipelineBasics, nil
}

func (c *controller) StopPipelinerun(ctx context.Context, pipelinerunID uint) (err error) {
	const op = "pipelinerun controller: stop pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	pipelinerun, err := c.pipelinerunMgr.GetByID(ctx, pipelinerunID)
	if err != nil {
		return errors.E(op, err)
	}
	if pipelinerun.Status != string(prmodels.StatusCreated) {
		return errors.E(op, http.StatusBadRequest, errors.ErrorCode("BadRequest"), "pipelinerun is already completed")
	}
	cluster, err := c.clusterMgr.GetByID(ctx, pipelinerun.ClusterID)
	if err != nil {
		return errors.E(op, err)
	}

	tektonClient, err := c.tektonFty.GetTekton(cluster.EnvironmentName)
	if err != nil {
		return errors.E(op, err)
	}

	return tektonClient.StopPipelineRun(ctx, pipelinerun.CIEventID)
}

func (c *controller) StopPipelinerunForCluster(ctx context.Context, clusterID uint) (err error) {
	const op = "pipelinerun controller: stop pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	if err != nil {
		return errors.E(op, err)
	}
	// get cluster latest builddeploy pipelinerun
	pipelinerun, err := c.pipelinerunMgr.GetLatestByClusterIDAndActions(ctx, clusterID, prmodels.ActionBuildDeploy)

	// if pipelinerun.Status is not created, ignore, and return success
	if pipelinerun.Status != string(prmodels.StatusCreated) {
		return nil
	}
	cluster, err := c.clusterMgr.GetByID(ctx, pipelinerun.ClusterID)
	if err != nil {
		return errors.E(op, err)
	}

	tektonClient, err := c.tektonFty.GetTekton(cluster.EnvironmentName)
	if err != nil {
		return errors.E(op, err)
	}

	return tektonClient.StopPipelineRun(ctx, pipelinerun.CIEventID)
}
