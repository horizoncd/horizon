package pipelinerun

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/q"
	appmanager "g.hz.netease.com/horizon/pkg/application/manager"
	"g.hz.netease.com/horizon/pkg/cluster/code"
	"g.hz.netease.com/horizon/pkg/cluster/gitrepo"
	clustermanager "g.hz.netease.com/horizon/pkg/cluster/manager"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/factory"
	"g.hz.netease.com/horizon/pkg/cluster/tekton/log"
	envmanager "g.hz.netease.com/horizon/pkg/environment/manager"
	perror "g.hz.netease.com/horizon/pkg/errors"
	prmanager "g.hz.netease.com/horizon/pkg/pipelinerun/manager"
	"g.hz.netease.com/horizon/pkg/pipelinerun/models"
	prmodels "g.hz.netease.com/horizon/pkg/pipelinerun/models"
	usermanager "g.hz.netease.com/horizon/pkg/user/manager"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

type Controller interface {
	GetPipelinerunLog(ctx context.Context, pipelinerunID uint) (*Log, error)
	GetClusterLatestLog(ctx context.Context, clusterID uint) (*Log, error)
	GetDiff(ctx context.Context, pipelinerunID uint) (*GetDiffResponse, error)
	Get(ctx context.Context, pipelinerunID uint) (*PipelineBasic, error)
	List(ctx context.Context, clusterID uint, canRollback bool, query q.Query) (int, []*PipelineBasic, error)
	StopPipelinerun(ctx context.Context, pipelinerunID uint) error
	StopPipelinerunForCluster(ctx context.Context, clusterID uint) error
}

type controller struct {
	pipelinerunMgr prmanager.Manager
	applicationMgr appmanager.Manager
	clusterMgr     clustermanager.Manager
	envMgr         envmanager.Manager
	tektonFty      factory.Factory
	commitGetter   code.GitGetter
	clusterGitRepo gitrepo.ClusterGitRepo
	userManager    usermanager.Manager
}

var _ Controller = (*controller)(nil)

func NewController(tektonFty factory.Factory, codeGetter code.GitGetter,
	clusterRepo gitrepo.ClusterGitRepo) Controller {
	return &controller{
		pipelinerunMgr: prmanager.Mgr,
		clusterMgr:     clustermanager.Mgr,
		envMgr:         envmanager.Mgr,
		tektonFty:      tektonFty,
		commitGetter:   codeGetter,
		applicationMgr: appmanager.Mgr,
		clusterGitRepo: clusterRepo,
		userManager:    usermanager.Mgr,
	}
}

type Log struct {
	LogChannel <-chan log.Log
	ErrChannel <-chan error

	LogBytes []byte
}

func (c *controller) GetPipelinerunLog(ctx context.Context, pipelinerunID uint) (_ *Log, err error) {
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

	// only builddeploy have logs
	if pr.Action != prmodels.ActionBuildDeploy {
		return nil, errors.E(op, fmt.Errorf("%v action has no log", pr.Action))
	}

	return c.getPipelinerunLog(ctx, pr, cluster, cluster.EnvironmentName)
}

func (c *controller) GetClusterLatestLog(ctx context.Context, clusterID uint) (_ *Log, err error) {
	const op = "pipelinerun controller: get cluster latest log"
	defer wlog.Start(ctx, op).StopPrint()

	pr, err := c.pipelinerunMgr.GetLatestByClusterIDAndAction(ctx, clusterID, prmodels.ActionBuildDeploy)
	if err != nil {
		return nil, errors.E(op, err)
	}
	if pr == nil {
		return nil, errors.E(op, fmt.Errorf("no builddeploy pipelinerun"))
	}

	cluster, err := c.clusterMgr.GetByID(ctx, clusterID)
	if err != nil {
		return nil, errors.E(op, err)
	}
	return c.getPipelinerunLog(ctx, pr, cluster, cluster.EnvironmentName)
}

func (c *controller) getPipelinerunLog(ctx context.Context, pr *prmodels.Pipelinerun, cluster *clustermodels.Cluster,
	environment string) (_ *Log, err error) {
	const op = "pipeline controller: get pipelinerun log"
	defer wlog.Start(ctx, op).StopPrint()

	// if pr.PrObject is empty, get pipelinerun log in k8s
	if pr.PrObject == "" {
		tektonClient, err := c.tektonFty.GetTekton(environment)
		if err != nil {
			return nil, perror.WithMessagef(err, "faild to get tekton for %s", environment)
		}

		logCh, errCh, err := tektonClient.GetPipelineRunLogByID(ctx, cluster.Name, cluster.ID, pr.ID)
		if err != nil {
			return nil, err
		}
		return &Log{
			LogChannel: logCh,
			ErrChannel: errCh,
		}, nil
	}

	// else, get log from s3
	tektonCollector, err := c.tektonFty.GetTektonCollector(environment)
	if err != nil {
		return nil, perror.WithMessagef(err, "faild to get tekton collector for %s", environment)
	}
	logBytes, err := tektonCollector.GetPipelineRunLog(ctx, pr.LogObject)
	if err != nil {
		return nil, perror.WithMessagef(err, "faild to get tekton collector for %s", environment)
	}
	return &Log{
		LogBytes: logBytes,
	}, nil
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
		pipelinerun.GitBranch != "" {
		commit, err := c.commitGetter.GetCommit(ctx, pipelinerun.GitURL, nil, &pipelinerun.GitCommit)
		if err != nil {
			return nil, err
		}
		var historyLink string
		if strings.HasPrefix(pipelinerun.GitURL, common.InternalGitSSHPrefix) {
			httpURL := common.InternalSSHToHTTPURL(pipelinerun.GitURL)
			historyLink = httpURL + common.CommitHistoryMiddle + pipelinerun.GitCommit
		}
		codeDiff = &CodeInfo{
			Branch:    pipelinerun.GitBranch,
			CommitID:  pipelinerun.GitCommit,
			CommitMsg: commit.Message,
			Link:      historyLink,
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
	pr, firstCanRollbackPipelinerun *models.Pipelinerun) (*PipelineBasic, error) {
	user, err := c.userManager.GetUserByID(ctx, pr.CreatedBy)
	if err != nil {
		return nil, err
	}

	canRollback := func() bool {
		// set the firstCanRollbackPipelinerun that cannot rollback
		if firstCanRollbackPipelinerun != nil && pr.ID == firstCanRollbackPipelinerun.ID {
			return false
		}
		return pr.Action != prmodels.ActionRestart && pr.Status == prmodels.ResultOK
	}()

	return &PipelineBasic{
		ID:               pr.ID,
		Title:            pr.Title,
		Description:      pr.Description,
		Action:           pr.Action,
		Status:           pr.Status,
		GitURL:           pr.GitURL,
		GitBranch:        pr.GitBranch,
		GitCommit:        pr.GitCommit,
		ImageURL:         pr.ImageURL,
		LastConfigCommit: pr.LastConfigCommit,
		ConfigCommit:     pr.ConfigCommit,
		CreatedAt:        pr.CreatedAt,
		StartedAt:        pr.StartedAt,
		FinishedAt:       pr.FinishedAt,
		CanRollback:      canRollback,
		CreatedBy: UserInfo{
			UserID:   pr.CreatedBy,
			UserName: user.Name,
		},
	}, nil
}

func (c *controller) ofPipelineBasics(ctx context.Context, prs []*models.Pipelinerun,
	firstCanRollbackPipelinerun *models.Pipelinerun) ([]*PipelineBasic, error) {
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
	if pipelinerun.Status != prmodels.ResultCreated {
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

	return tektonClient.StopPipelineRun(ctx, cluster.Name, cluster.ID, pipelinerunID)
}

func (c *controller) StopPipelinerunForCluster(ctx context.Context, clusterID uint) (err error) {
	const op = "pipelinerun controller: stop pipelinerun"
	defer wlog.Start(ctx, op).StopPrint()

	if err != nil {
		return errors.E(op, err)
	}
	// get cluster latest builddeploy pipelinerun
	pipelinerun, err := c.pipelinerunMgr.GetLatestByClusterIDAndAction(ctx, clusterID, prmodels.ActionBuildDeploy)

	// if pipelinerun.Status is not created, ignore, and return success
	if pipelinerun.Status != prmodels.ResultCreated {
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

	return tektonClient.StopPipelineRun(ctx, cluster.Name, cluster.ID, pipelinerun.ID)
}
