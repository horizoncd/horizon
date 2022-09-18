package gitrepo

import (
	"context"
	"encoding/json"
	goerrors "errors"
	"fmt"
	"strings"
	"sync"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	"g.hz.netease.com/horizon/pkg/application/models"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	perror "g.hz.netease.com/horizon/pkg/errors"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	tagmodels "g.hz.netease.com/horizon/pkg/tag/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/templaterepo"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/mergemap"
	timeutil "g.hz.netease.com/horizon/pkg/util/time"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
	kyaml "sigs.k8s.io/yaml"
)

/*
music-cloud-native
	  │
      ├── applications                 -- 应用配置 group
      │    └── app1                    -- 应用 group
      │         ├── application.yaml   -- 部署模板参数
      │         └── pipeline.yaml      -- 流水线参数
      │
      └── clusters                                            -- 集群配置 group
      			└──	app1                                      -- 应用 group
                    └──Cluster-1                              -- 集群 repo
                              ├── Chart.yaml
                              ├── application.yaml            -- 用户实际数据
							  ├── tags.yaml                   -- tags数据
                              ├── sre                         -- sre目录
                              │     └── sre.yaml              -- sre values数据
                              ├── system
                              │     ├── horizon.yaml          -- 基础数据
                              │     ├── restart.yaml          -- 重启时间
                              │     └── env.yaml              -- 环境相关数据
                              └── pipeline
                                    ├── pipeline.yaml         -- pipeline模板参数
                                    └── pipeline-output.yaml  -- pipeline输出
*/

const (
	// _branchMaster is the main branch
	_branchMaster = "master"
	// _branchGitops is the gitops branch, values updated in this branch, then merge into the _branchMaster
	_branchGitops = "gitops"

	// fileName
	_filePathChart          = "Chart.yaml"
	_filePathApplication    = "application.yaml"
	_filePathTags           = "tags.yaml"
	_filePathSRE            = "sre/sre.yaml"
	_filePathBase           = "system/horizon.yaml"
	_filePathEnv            = "system/env.yaml"
	_filePathRestart        = "system/restart.yaml"
	_filePathPipeline       = "pipeline/pipeline.yaml"
	_filePathPipelineOutput = "pipeline/pipeline-output.yaml"
	_filePathManifest       = "manifest.yaml"

	// value namespace
	_envValueNamespace  = "env"
	_baseValueNamespace = "horizon"

	_mergeRequestStateOpen = "opened"
)

const (
	VersionV2           = "0.0.2"
	PipelineValueParent = "pipeline"
)

var ErrPipelineOutputEmpty = goerrors.New("PipelineOutput is empty")

type BaseParams struct {
	ClusterID           uint
	Cluster             string
	PipelineJSONBlob    map[string]interface{}
	ApplicationJSONBlob map[string]interface{}
	TemplateRelease     *trmodels.TemplateRelease
	Application         *models.Application
	Environment         string
	RegionEntity        *regionmodels.RegionEntity
	Namespace           string

	Version string
}

type Manifest struct {
	// TODO(encode the template info into manifest),currently only the Version
	Version string `yaml:"version"`
}

type CreateClusterParams struct {
	*BaseParams
	Tags  []*tagmodels.Tag
	Image string
}

type UpdateClusterParams struct {
	*BaseParams
}

type RepoInfo struct {
	GitRepoSSHURL string
	ValueFiles    []string
}

type ClusterFiles struct {
	PipelineJSONBlob    map[string]interface{}
	ApplicationJSONBlob map[string]interface{}
	Manifest            map[string]interface{}
}

type ClusterValueFile struct {
	FileName string
	Content  map[interface{}]interface{}
}

type ClusterCommit struct {
	Master string
	Gitops string
}

type ClusterGitRepo interface {
	GetCluster(ctx context.Context, application, cluster, templateName string) (*ClusterFiles, error)
	GetClusterValueFiles(ctx context.Context,
		application, cluster string) ([]ClusterValueFile, error)
	CreateCluster(ctx context.Context, params *CreateClusterParams) error
	UpdateCluster(ctx context.Context, params *UpdateClusterParams) error
	DeleteCluster(ctx context.Context, application, cluster string, clusterID uint) error
	HardDeleteCluster(ctx context.Context, application, cluster string) error
	// CompareConfig compare config of `from` commit with `to` commit.
	// if `from` or `to` is nil, compare the master branch with gitops branch
	CompareConfig(ctx context.Context, application, cluster string, from, to *string) (string, error)
	// MergeBranch merge gitops branch to master branch, and return master branch's newest commit
	MergeBranch(ctx context.Context, application, cluster string, prID uint) (string, error)
	GetPipelineOutput(ctx context.Context, application, cluster string, template string) (*PipelineOutput, error)
	UpdatePipelineOutput(ctx context.Context, application, cluster, template string,
		PipelineOutput interface{}) (string, error)
	// UpdateRestartTime update restartTime in git repo for restart
	// TODO(gjq): some template cannot restart, for example serverless, how to do it ?
	GetRestartTime(ctx context.Context, application, cluster string,
		template string) (string, error)
	UpdateRestartTime(ctx context.Context, application, cluster, template string) (string, error)
	GetConfigCommit(ctx context.Context, application, cluster string) (*ClusterCommit, error)
	GetRepoInfo(ctx context.Context, application, cluster string) *RepoInfo
	GetEnvValue(ctx context.Context, application, cluster, templateName string) (*EnvValue, error)
	Rollback(ctx context.Context, application, cluster, commit string) (string, error)
	UpdateTags(ctx context.Context, application, cluster, templateName string,
		tags []*tagmodels.Tag) error
}

type clusterGitRepo struct {
	gitlabLib       gitlablib.Interface
	clusterRepoConf *gitlabconf.Repo
	templateRepo    templaterepo.TemplateRepo
}

func NewClusterGitlabRepo(ctx context.Context, gitlabRepoConfig gitlabconf.RepoConfig,
	templateRepo templaterepo.TemplateRepo, gitlabLib gitlablib.Interface) (ClusterGitRepo, error) {
	return &clusterGitRepo{
		gitlabLib:       gitlabLib,
		clusterRepoConf: gitlabRepoConfig.Cluster,
		templateRepo:    templateRepo,
	}, nil
}

func (g *clusterGitRepo) GetCluster(ctx context.Context,
	application, cluster, templateName string) (_ *ClusterFiles, err error) {
	const op = "cluster git repo: get cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get template and pipeline from gitlab
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	var applicationBytes, pipelineBytes, manifestBytes []byte
	var err1, err2, err3 error

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		pipelineBytes, err1 = g.gitlabLib.GetFile(ctx, pid, _branchGitops, _filePathPipeline)
		if err1 != nil {
			return
		}
		pipelineBytes, err1 = kyaml.YAMLToJSON(pipelineBytes)
		if err1 != nil {
			err1 = perror.Wrap(herrors.ErrParamInvalid, err1.Error())
		}
	}()
	go func() {
		defer wg.Done()
		applicationBytes, err2 = g.gitlabLib.GetFile(ctx, pid, _branchGitops, _filePathApplication)
		if err2 != nil {
			return
		}
		applicationBytes, err2 = kyaml.YAMLToJSON(applicationBytes)
		if err2 != nil {
			err2 = perror.Wrap(herrors.ErrParamInvalid, err2.Error())
		}
	}()
	go func() {
		defer wg.Done()
		manifestBytes, err3 = g.gitlabLib.GetFile(ctx, pid, _branchGitops, _filePathManifest)
		if err3 != nil {
			return
		}
		manifestBytes, err3 = kyaml.YAMLToJSON(manifestBytes)
		if err3 != nil {
			return
		}
	}()
	wg.Wait()

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			if _, ok := perror.Cause(err3).(*herrors.HorizonErrNotFound); !ok {
				return nil, err
			}
		}
	}

	var pipelineJSONBlobWithTemplate, applicationJSONBlobWithTemplate map[string]map[string]interface{}
	if pipelineBytes != nil {
		if err := json.Unmarshal(pipelineBytes, &pipelineJSONBlobWithTemplate); err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
	}
	if applicationBytes != nil {
		if err := json.Unmarshal(applicationBytes, &applicationJSONBlobWithTemplate); err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
	}
	var manifest = Manifest{}
	if manifestBytes != nil {
		err = json.Unmarshal(manifestBytes, &manifest)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
	}

	pipelineJSONBlob, err := func() (map[string]interface{}, error) {
		var pipelineValueParentName string
		if manifest.Version == "" {
			pipelineValueParentName = templateName
		} else {
			pipelineValueParentName = PipelineValueParent
		}
		jsonBlob, ok := pipelineJSONBlobWithTemplate[pipelineValueParentName]
		if !ok {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"pipeline value parent prefix not found, prefix = %s ", pipelineValueParentName)
		}
		return jsonBlob, nil
	}()
	if err != nil {
		return nil, err
	}

	applicationJSONBlob, ok := applicationJSONBlobWithTemplate[templateName]
	if !ok {
		return nil, perror.Wrapf(herrors.ErrParamInvalid,
			"template value parent prefix not found, prefix = %s ", templateName)
	}

	manifestJSONBlob, err := func() (map[string]interface{}, error) {
		if manifestBytes == nil {
			return nil, nil
		}
		var jsonBlob map[string]interface{}
		err := json.Unmarshal(manifestBytes, &jsonBlob)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"manifest Unmarshal error, value = %s ", string(manifestBytes))
		}
		return jsonBlob, nil
	}()
	if err != nil {
		return nil, err
	}

	return &ClusterFiles{
		PipelineJSONBlob:    pipelineJSONBlob,
		ApplicationJSONBlob: applicationJSONBlob,
		Manifest:            manifestJSONBlob,
	}, nil
}

func (g *clusterGitRepo) GetClusterValueFiles(ctx context.Context,
	application, cluster string) (_ []ClusterValueFile, err error) {
	const op = "cluster git repo: get cluster value files"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get  value file from git
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	cases := []struct {
		retBytes []byte
		err      error
		fileName string
	}{
		{
			fileName: _filePathBase,
		}, {
			fileName: _filePathEnv,
		}, {
			fileName: _filePathApplication,
		}, {
			fileName: _filePathTags,
		}, {
			fileName: _filePathSRE,
		},
	}

	var wg sync.WaitGroup
	wg.Add(len(cases))
	for i := 0; i < len(cases); i++ {
		go func(index int) {
			defer wg.Done()
			cases[index].retBytes, cases[index].err = g.gitlabLib.GetFile(ctx, pid,
				_branchMaster, cases[index].fileName)
			if cases[index].err != nil {
				log.Warningf(ctx, "get file %s error, err = %s", cases[index].fileName, cases[index].err.Error())
			}
		}(i)
	}
	wg.Wait()

	for i := 0; i < len(cases); i++ {
		if cases[i].err != nil {
			if _, ok := perror.Cause(cases[i].err).(*herrors.HorizonErrNotFound); !ok {
				log.Errorf(ctx, "get cluster value file error, err = %s", cases[i].err.Error())
				return nil, cases[i].err
			}
		}
	}

	// 2. check yaml format ok
	var clusterValueFiles []ClusterValueFile
	for _, oneCase := range cases {
		if oneCase.err != nil {
			if _, ok := perror.Cause(oneCase.err).(*herrors.HorizonErrNotFound); ok {
				continue
			}
		}
		var out map[interface{}]interface{}
		err = yaml.Unmarshal(oneCase.retBytes, &out)
		if err != nil {
			err = perror.Wrapf(herrors.ErrParamInvalid, "yaml Unmarshal err, file = %s", oneCase.fileName)
			break
		}

		clusterValueFiles = append(clusterValueFiles, ClusterValueFile{
			FileName: oneCase.fileName,
			Content:  out,
		})
	}
	if err != nil {
		return nil, err
	}

	return clusterValueFiles, nil
}

func (g *clusterGitRepo) CreateCluster(ctx context.Context, params *CreateClusterParams) (err error) {
	const op = "cluster git repo: create cluster"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	// 1. create application group if necessary
	var appGroup *gitlab.Group
	appGroup, err = g.gitlabLib.GetGroup(ctx, fmt.Sprintf("%v/%v", g.clusterRepoConf.Parent.Path, params.Application.Name))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
		appGroup, err = g.gitlabLib.CreateGroup(ctx, params.Application.Name,
			params.Application.Name, &g.clusterRepoConf.Parent.ID)
		if err != nil {
			return errors.E(op, err)
		}
	}

	// 3. create cluster repo under appGroup
	if _, err := g.gitlabLib.CreateProject(ctx, params.Cluster, appGroup.ID); err != nil {
		return err
	}

	// 3. create gitops branch from master
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, params.Application.Name, params.Cluster)
	if _, err := g.gitlabLib.CreateBranch(ctx, pid, _branchGitops, _branchMaster); err != nil {
		return err
	}

	// 4. write files to repo, to gitops branch
	var applicationYAML, pipelineYAML, baseValueYAML []byte
	var envValueYAML, sreValueYAML, chartYAML, restartYAML, tagsYAML, manifestValueYAML []byte
	var err1, err2, err3, err4, err5, err6, err7, err8, err9 error

	if params.BaseParams.ApplicationJSONBlob != nil {
		marshal(&applicationYAML, &err1, g.assembleApplicationValue(params.BaseParams))
	}
	if params.BaseParams.PipelineJSONBlob != nil {
		marshal(&pipelineYAML, &err2, g.assemblePipelineValue(params.BaseParams))
	}
	marshal(&baseValueYAML, &err3, g.assembleBaseValue(params.BaseParams))
	marshal(&envValueYAML, &err4, g.assembleEnvValue(params.BaseParams))
	marshal(&sreValueYAML, &err5, g.assembleSREValue(params))
	marshal(&restartYAML, &err7, assembleRestart(params.TemplateRelease.ChartName))
	marshal(&tagsYAML, &err8, assembleTags(params.TemplateRelease.ChartName, params.Tags))
	if params.BaseParams.Version != "" {
		marshal(&manifestValueYAML, &err9, Manifest{Version: params.BaseParams.Version})
	}

	chart, err := g.assembleChart(params.BaseParams)
	if err != nil {
		return err
	}
	marshal(&chartYAML, &err6, chart)

	for _, err := range []error{err1, err2, err3, err4, err5, err6, err7, err8} {
		if err != nil {
			return err
		}
	}
	actions := func() []gitlablib.CommitAction {
		gitActions := []gitlablib.CommitAction{
			{
				Action:   gitlablib.FileCreate,
				FilePath: _filePathTags,
				Content:  string(tagsYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: _filePathBase,
				Content:  string(baseValueYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: _filePathEnv,
				Content:  string(envValueYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: _filePathSRE,
				Content:  string(sreValueYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: _filePathChart,
				Content:  string(chartYAML),
			},
			// create _filePathRestart file first
			{
				Action:   gitlablib.FileCreate,
				FilePath: _filePathRestart,
				Content:  string(restartYAML),
			},
		}

		if applicationYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileCreate,
				FilePath: _filePathApplication,
				Content:  string(applicationYAML),
			})
		}
		if pipelineYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileCreate,
				FilePath: _filePathPipeline,
				Content:  string(pipelineYAML),
			})
		}
		if manifestValueYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileCreate,
				FilePath: _filePathManifest,
				Content:  "",
			})
		}
		return gitActions
	}()

	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "create cluster",
		Cluster:  angular.StringPtr(params.Cluster),
	}, struct {
		Application map[string]interface{} `json:"application"`
		Pipeline    map[string]interface{} `json:"pipeline"`
	}{
		Application: params.ApplicationJSONBlob,
		Pipeline:    params.PipelineJSONBlob,
	})

	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchGitops, commitMsg, nil, actions); err != nil {
		return err
	}

	return nil
}

func (g *clusterGitRepo) UpdateCluster(ctx context.Context, params *UpdateClusterParams) (err error) {
	const op = "cluster git repo: update cluster"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	// 1. write files to repo
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, params.Application.Name, params.Cluster)
	var applicationYAML, pipelineYAML, baseValueYAML, envValueYAML, chartYAML []byte
	var err1, err2, err3, err4, err5 error
	if params.Application != nil {
		marshal(&applicationYAML, &err1, g.assembleApplicationValue(params.BaseParams))
	}
	if params.PipelineJSONBlob != nil {
		marshal(&pipelineYAML, &err2, g.assemblePipelineValue(params.BaseParams))
	}
	marshal(&baseValueYAML, &err3, g.assembleBaseValue(params.BaseParams))
	chart, err := g.assembleChart(params.BaseParams)
	if err != nil {
		return err
	}
	marshal(&chartYAML, &err4, chart)
	if params.RegionEntity != nil {
		marshal(&envValueYAML, &err5, g.assembleEnvValue(params.BaseParams))
	}
	// TODO currently not support update manifest

	for _, err := range []error{err1, err2, err3, err4, err5} {
		if err != nil {
			return err
		}
	}

	actions := func() []gitlablib.CommitAction {
		gitActions := []gitlablib.CommitAction{
			{
				Action:   gitlablib.FileUpdate,
				FilePath: _filePathBase,
				Content:  string(baseValueYAML),
			}, {
				Action:   gitlablib.FileUpdate,
				FilePath: _filePathChart,
				Content:  string(chartYAML),
			},
		}
		if applicationYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileUpdate,
				FilePath: _filePathApplication,
				Content:  string(applicationYAML),
			})
		}
		if pipelineYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileUpdate,
				FilePath: _filePathPipeline,
				Content:  string(pipelineYAML),
			})
		}
		if envValueYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileUpdate,
				FilePath: _filePathEnv,
				Content:  string(envValueYAML),
			})
		}
		return gitActions
	}()

	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "update cluster",
		Cluster:  angular.StringPtr(params.Cluster),
	}, struct {
		Application map[string]interface{} `json:"application"`
		Pipeline    map[string]interface{} `json:"pipeline"`
	}{
		Application: params.ApplicationJSONBlob,
		Pipeline:    params.PipelineJSONBlob,
	})
	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchGitops, commitMsg, nil, actions); err != nil {
		return err
	}

	return nil
}

func (g *clusterGitRepo) DeleteCluster(ctx context.Context, application, cluster string, clusterID uint) (err error) {
	const op = "cluster git repo: delete cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. create application group if necessary
	_, err = g.gitlabLib.GetGroup(ctx, fmt.Sprintf("%v/%v", g.clusterRepoConf.RecyclingParent.Path, application))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
		_, err = g.gitlabLib.CreateGroup(ctx, application, application, &g.clusterRepoConf.RecyclingParent.ID)
		if err != nil {
			return err
		}
	}

	// 1. delete gitlab project
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	// 1.1 edit project's name and path to {cluster}-{clusterID}
	newName := fmt.Sprintf("%v-%d", cluster, clusterID)
	newPath := newName
	if err := g.gitlabLib.EditNameAndPathForProject(ctx, pid, &newName, &newPath); err != nil {
		return err
	}

	// 1.2 transfer project to RecyclingParent
	newPid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, newPath)
	return g.gitlabLib.TransferProject(ctx, newPid,
		fmt.Sprintf("%v/%v", g.clusterRepoConf.RecyclingParent.Path, application))
}

func (g *clusterGitRepo) HardDeleteCluster(ctx context.Context, application,
	cluster string) (err error) {
	const op = "cluster git repo: hard delete cluster"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	return g.gitlabLib.DeleteProject(ctx, pid)
}

func (g *clusterGitRepo) CompareConfig(ctx context.Context, application,
	cluster string, from, to *string) (_ string, err error) {
	const op = "cluster git repo: compare config"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	var compare *gitlab.Compare
	if from == nil || to == nil {
		compare, err = g.gitlabLib.Compare(ctx, pid, _branchMaster, _branchGitops, nil)
	} else {
		compare, err = g.gitlabLib.Compare(ctx, pid, *from, *to, nil)
	}
	if err != nil {
		return "", err
	}
	if compare.Diffs == nil {
		return "", nil
	}
	diffStr := ""
	for _, diff := range compare.Diffs {
		diffStr += "--- " + diff.OldPath + "\n"
		diffStr += "+++ " + diff.NewPath + "\n"
		diffStr += diff.Diff + "\n"
	}
	return diffStr, nil
}

func (g *clusterGitRepo) MergeBranch(ctx context.Context, application, cluster string,
	prID uint) (_ string, err error) {
	mergeCommitMsg := "git merge gitops"
	removeSourceBranch := false
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	title := fmt.Sprintf("git merge %v, id = %d", _branchGitops, prID)

	var mr *gitlab.MergeRequest
	mrs, err := g.gitlabLib.ListMRs(ctx, pid, _branchGitops, _branchMaster, _mergeRequestStateOpen)
	if err != nil {
		return "", perror.WithMessage(err, "failed to list merge requests")
	}
	if len(mrs) > 0 {
		// merge old mr when it is existed, because given specified source and target, gitlab only allows 1 mr to exist
		mr = mrs[0]

		// close the redundant mrs
		// gitlab has a bug for when concurrency create merge request(will exist 2 more merge request for the same
		// (source,target), caused we can't merge anymore)
		if len(mrs) >= 2 {
			log.Warningf(ctx, "there %d mrs for (src:%s, des:%s), here will kill redundant mrs",
				len(mrs), _branchGitops, _branchMaster)
			for i := 1; i < len(mrs); i++ {
				_, err := g.gitlabLib.CloseMR(ctx, pid, mrs[i].IID)
				if err != nil {
					return "", err
				}
			}
		}
	} else {
		// create new mr
		mr, err = g.gitlabLib.CreateMR(ctx, pid, _branchGitops, _branchMaster, title)
		if err != nil {
			return "", perror.WithMessage(err, "failed to create new merge request")
		}
	}

	mr, err = g.gitlabLib.AcceptMR(ctx, pid, mr.IID, &mergeCommitMsg, &removeSourceBranch)
	if err != nil {
		return "", perror.WithMessage(err, "failed to accept merge request")
	}
	return mr.MergeCommitSHA, nil
}

func (g *clusterGitRepo) GetManifest(ctx context.Context, application, cluster string) (*Manifest, error) {
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	content, err := g.gitlabLib.GetFile(ctx, pid, _branchGitops, _filePathManifest)
	if err != nil {
		return nil, err
	}
	manifest, err := func() (*Manifest, error) {
		manifestOutputBytes, err := kyaml.YAMLToJSON(content)
		if err != nil {
			return nil, err
		}
		ret := Manifest{}
		if err = json.Unmarshal(manifestOutputBytes, &ret); err != nil {
			return nil, err
		}
		return &ret, err
	}()
	if err != nil {
		return nil, err
	}
	return manifest, nil
}

func (g *clusterGitRepo) GetPipelineOutput(ctx context.Context, application, cluster string,
	template string) (*PipelineOutput, error) {
	ret := make(map[string]*PipelineOutput)
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	content, err := g.gitlabLib.GetFile(ctx, pid, _branchGitops, _filePathPipelineOutput)
	if err != nil {
		return nil, perror.WithMessage(err, "failed to get gitlab file")
	}

	pipelineOutputBytes, err := kyaml.YAMLToJSON(content)
	if err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	if err := json.Unmarshal(pipelineOutputBytes, &ret); err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	pipelineOutput, ok := ret[template]
	if !ok {
		return nil, perror.Wrapf(herrors.ErrPipelineOutputEmpty, "no template in pipelineOutput.yaml")
	}

	if pipelineOutput.Git == nil {
		pipelineOutput.Git = &Git{}
	}
	return pipelineOutput, nil
}

func (g *clusterGitRepo) getPipelineOutPut(ctx context.Context,
	application, cluster string) (map[string]map[string]interface{}, error) {
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	content, err := g.gitlabLib.GetFile(ctx, pid, _branchGitops, _filePathPipelineOutput)
	if err != nil {
		return nil, err
	}

	pipelineOutput, err := func() (map[string]map[string]interface{}, error) {
		pipelineOutputJSONBytes, err := kyaml.YAMLToJSON(content)
		if err != nil {
			return nil, err
		}
		var ret map[string]map[string]interface{}
		if err = json.Unmarshal(pipelineOutputJSONBytes, &ret); err != nil {
			return nil, err
		}
		return ret, nil
	}()
	if err != nil {
		return nil, perror.Wrap(herrors.ErrPipelineOutPut, err.Error())
	}

	return pipelineOutput, nil
}

func (g *clusterGitRepo) UpdatePipelineOutput(ctx context.Context, application, cluster, template string,
	pipelineOutPut interface{}) (commitID string, err error) {
	const op = "cluster git repo: update pipeline output"
	defer wlog.Start(ctx, op).StopPrint()

	pipelineOutPutInternalFormat, err := func() (map[string]interface{}, error) {
		bytes, err := json.Marshal(pipelineOutPut)
		if err != nil {
			return nil, err
		}
		var internalFormat map[string]interface{}
		if err := json.Unmarshal(bytes, &internalFormat); err != nil {
			return nil, err
		}
		return internalFormat, nil
	}()
	if err != nil {
		return "", err
	}

	pipelineValueParent, err := func() (string, error) {
		manifest, err := g.GetManifest(ctx, application, cluster)
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
				return "", err
			}
		}
		if manifest == nil || manifest.Version == "" {
			return template, nil
		}
		return PipelineValueParent, nil
	}()
	if err != nil {
		return "", err
	}

	newPipelineOutputContent := make(map[string]interface{})
	var PipelineOutPutFileExist bool
	currentPipelineOutputContent, err := g.getPipelineOutPut(ctx, application, cluster)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return "", err
		}
		newPipelineOutputContent[pipelineValueParent] = pipelineOutPutInternalFormat
	} else {
		PipelineOutPutFileExist = true
		// if current exist, just patch it
		newPipelineOutputContent[pipelineValueParent], err =
			mergemap.Merge(currentPipelineOutputContent[pipelineValueParent], pipelineOutPutInternalFormat)
		if err != nil {
			return "", perror.Wrap(herrors.ErrPipelineOutPut, err.Error())
		}
	}

	newPipelineOutPutBytes, err := kyaml.Marshal(newPipelineOutputContent)
	if err != nil {
		return "", perror.Wrap(herrors.ErrPipelineOutPut, err.Error())
	}

	actions := []gitlablib.CommitAction{
		{
			Action: func() gitlablib.FileAction {
				if PipelineOutPutFileExist {
					return gitlablib.FileUpdate
				}
				return gitlablib.FileCreate
			}(),
			FilePath: _filePathPipelineOutput,
			Content:  string(newPipelineOutPutBytes),
		},
	}

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return "", err
	}
	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "deploy cluster",
		Cluster:  angular.StringPtr(cluster),
	}, pipelineOutPut)

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	commit, err := g.gitlabLib.WriteFiles(ctx, pid, _branchGitops, commitMsg, nil, actions)
	if err != nil {
		return "", perror.WithMessage(err, "failed to write gitlab files")
	}
	return commit.ID, nil
}

func (g *clusterGitRepo) GetRestartTime(ctx context.Context, application, cluster string,
	template string) (string, error) {
	ret := make(map[string]map[string]string)
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	content, err := g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathRestart)
	if err != nil {
		return "", perror.WithMessage(err, "failed to get gitlab file")
	}

	restartBytes, err := kyaml.YAMLToJSON(content)
	if err != nil {
		return "", perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	if err := json.Unmarshal(restartBytes, &ret); err != nil {
		return "", perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	restartStr, ok := ret[template]["restartTime"]
	if !ok {
		return "", perror.Wrapf(herrors.ErrRestartFileEmpty, "no template in restart.yaml")
	}

	return restartStr, nil
}

func (g *clusterGitRepo) UpdateRestartTime(ctx context.Context,
	application, cluster, template string) (_ string, err error) {
	const op = "cluster git repo: update restartTime"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return "", err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	var restartYAML []byte
	var err1 error
	marshal(&restartYAML, &err1, assembleRestart(template))
	if err1 != nil {
		return "", err1
	}

	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathRestart,
			Content:  string(restartYAML),
		},
	}

	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "restart cluster",
		Cluster:  angular.StringPtr(cluster),
	}, nil)

	// update in _branchMaster directly
	commit, err := g.gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions)
	if err != nil {
		return "", err
	}

	return commit.ID, nil
}

func (g *clusterGitRepo) GetConfigCommit(ctx context.Context,
	application, cluster string) (_ *ClusterCommit, err error) {
	const op = "cluster git repo: get config commit"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	var branchMaster, branchGitops *gitlab.Branch
	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		branchMaster, err1 = g.gitlabLib.GetBranch(ctx, pid, _branchMaster)
	}()
	go func() {
		defer wg.Done()
		branchGitops, err2 = g.gitlabLib.GetBranch(ctx, pid, _branchGitops)
	}()
	wg.Wait()

	for _, err := range []error{err1, err2} {
		if err != nil {
			return nil, err
		}
	}

	return &ClusterCommit{
		Master: branchMaster.Commit.ID,
		Gitops: branchGitops.Commit.ID,
	}, nil
}

func (g *clusterGitRepo) GetRepoInfo(ctx context.Context, application, cluster string) *RepoInfo {
	return &RepoInfo{
		GitRepoSSHURL: fmt.Sprintf("%v/%v/%v/%v.git",
			g.gitlabLib.GetSSHURL(ctx), g.clusterRepoConf.Parent.Path, application, cluster),
		ValueFiles: []string{_filePathApplication, _filePathPipelineOutput,
			_filePathEnv, _filePathBase, _filePathTags, _filePathRestart, _filePathSRE},
	}
}

func (g *clusterGitRepo) GetEnvValue(ctx context.Context,
	application, cluster, templateName string) (_ *EnvValue, err error) {
	const op = "cluster git repo: get config commit"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	bytes, err := g.gitlabLib.GetFile(ctx, pid, _branchGitops, _filePathEnv)
	if err != nil {
		return nil, err
	}

	var envMap map[string]map[string]*EnvValue

	if err := yaml.Unmarshal(bytes, &envMap); err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	return envMap[templateName][_envValueNamespace], nil
}

func (g *clusterGitRepo) Rollback(ctx context.Context, application, cluster, commit string) (_ string, err error) {
	const op = "cluster git repo: rollback"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return "", err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	// 1. get cluster files of the specified commit
	// the files contains: _filePathPipeline, _filePathApplication, _filePathBase, _filePathPipelineOutput
	var err1, err2, err3, err4, err5 error
	var applicationBytes, pipelineBytes, baseValueBytes, pipelineOutPutBytes, tagsBytes []byte
	var wgReadFile sync.WaitGroup
	wgReadFile.Add(5)
	readFile := func(b *[]byte, err *error, filePath string) {
		defer wgReadFile.Done()
		bytes, e := g.gitlabLib.GetFile(ctx, pid, commit, filePath)
		*b = bytes
		*err = e
	}
	go readFile(&pipelineBytes, &err1, _filePathPipeline)
	go readFile(&applicationBytes, &err2, _filePathApplication)
	go readFile(&baseValueBytes, &err3, _filePathBase)
	go readFile(&pipelineOutPutBytes, &err4, _filePathPipelineOutput)
	go readFile(&tagsBytes, &err5, _filePathTags)
	wgReadFile.Wait()

	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return "", err
		}
	}

	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathPipeline,
			Content:  string(pipelineBytes),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathApplication,
			Content:  string(applicationBytes),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathBase,
			Content:  string(baseValueBytes),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathPipelineOutput,
			Content:  string(pipelineOutPutBytes),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathTags,
			Content:  string(tagsBytes),
		},
	}

	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "rollback cluster",
		Cluster:  angular.StringPtr(cluster),
	}, struct {
		Commit string `json:"commit"`
	}{
		Commit: commit,
	})

	newCommit, err := g.gitlabLib.WriteFiles(ctx, pid, _branchGitops, commitMsg, nil, actions)
	if err != nil {
		return "", err
	}

	return newCommit.ID, nil
}

func (g *clusterGitRepo) UpdateTags(ctx context.Context, application, cluster, templateName string,
	tags []*tagmodels.Tag) (err error) {
	const op = "cluster git repo: update tags"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	var tagsYAML []byte
	marshal(&tagsYAML, &err, assembleTags(templateName, tags))
	if err != nil {
		return err
	}

	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathTags,
			Content:  string(tagsYAML),
		},
	}

	type Tag struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "update tags",
		Cluster:  angular.StringPtr(cluster),
	}, struct {
		Tags []*Tag `json:"tags"`
	}{
		Tags: func(tags []*tagmodels.Tag) []*Tag {
			clusterTags := make([]*Tag, 0, len(tags))
			for _, tag := range tags {
				clusterTags = append(clusterTags, &Tag{
					Key:   tag.Key,
					Value: tag.Value,
				})
			}
			return clusterTags
		}(tags),
	})

	_, err = g.gitlabLib.WriteFiles(ctx, pid, _branchGitops, commitMsg, nil, actions)
	if err != nil {
		return err
	}

	return nil
}

// assembleApplicationValue assemble application.yaml data
func (g *clusterGitRepo) assembleApplicationValue(params *BaseParams) map[string]map[string]interface{} {
	ret := make(map[string]map[string]interface{})
	ret[params.TemplateRelease.ChartName] = params.ApplicationJSONBlob
	return ret
}

// assembleApplicationValue assemble pipeline.yaml data
func (g *clusterGitRepo) assemblePipelineValue(params *BaseParams) map[string]map[string]interface{} {
	ret := make(map[string]map[string]interface{})
	// the default version prefix pipeline value with template ChartName
	if params.Version == "" {
		ret[params.TemplateRelease.ChartName] = params.PipelineJSONBlob
	} else {
		ret[PipelineValueParent] = params.PipelineJSONBlob
	}
	return ret
}

// assembleSreValue assemble sre value data
func (g *clusterGitRepo) assembleSREValue(params *CreateClusterParams) map[string]interface{} {
	ret := make(map[string]interface{})
	ret[params.TemplateRelease.ChartName] = make(map[string]string)
	return ret
}

type EnvValue struct {
	Environment   string `yaml:"environment"`
	Region        string `yaml:"region"`
	Namespace     string `yaml:"namespace"`
	BaseRegistry  string `yaml:"baseRegistry"`
	IngressDomain string `yaml:"ingressDomain"`
}

func getNamespace(params *BaseParams) string {
	if params.Namespace != "" {
		return params.Namespace
	}
	return fmt.Sprintf("%v-%v", params.Environment, params.Application.GroupID)
}

func (g *clusterGitRepo) assembleEnvValue(params *BaseParams) map[string]map[string]*EnvValue {
	envMap := make(map[string]*EnvValue)
	envMap[_envValueNamespace] = &EnvValue{
		Environment: params.Environment,
		Region:      params.RegionEntity.Name,
		Namespace:   getNamespace(params),
		BaseRegistry: strings.TrimPrefix(strings.TrimPrefix(
			params.RegionEntity.Harbor.Server, "https://"), "http://"),
		IngressDomain: params.RegionEntity.IngressDomain,
	}

	ret := make(map[string]map[string]*EnvValue)
	ret[params.TemplateRelease.ChartName] = envMap
	return ret
}

type BaseValue struct {
	Application string             `yaml:"application"`
	ClusterID   uint               `yaml:"clusterID"`
	Cluster     string             `yaml:"cluster"`
	Template    *BaseValueTemplate `yaml:"template"`
	Priority    string             `yaml:"priority"`
}

type PipelineOutput struct {
	Image *string `yaml:"image,omitempty" json:"image,omitempty"`
	Git   *Git    `yaml:"git,omitempty" json:"git,omitempty"`
}

type Git struct {
	URL      *string `yaml:"url,omitempty" json:"url,omitempty"`
	CommitID *string `yaml:"commitID,omitempty" json:"commitID,omitempty"`
	Branch   *string `yaml:"branch,omitempty" json:"branch,omitempty"`
	Tag      *string `yaml:"tag,omitempty" json:"tag,omitempty"`
}

type BaseValueTemplate struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

// assembleBaseValue assemble base value. return a map, key is template name,
// and value is a map which key is "horizon", and value is *BaseValue
func (g *clusterGitRepo) assembleBaseValue(params *BaseParams) map[string]map[string]*BaseValue {
	baseMap := make(map[string]*BaseValue)
	baseMap[_baseValueNamespace] = &BaseValue{
		Application: params.Application.Name,
		ClusterID:   params.ClusterID,
		Cluster:     params.Cluster,
		Template: &BaseValueTemplate{
			Name:    params.TemplateRelease.TemplateName,
			Release: params.TemplateRelease.ChartVersion,
		},
		Priority: string(params.Application.Priority),
	}

	ret := make(map[string]map[string]*BaseValue)
	ret[params.TemplateRelease.ChartName] = baseMap
	return ret
}

type Chart struct {
	APIVersion   string       `yaml:"apiVersion"`
	Name         string       `yaml:"name"`
	Version      string       `yaml:"version"`
	Dependencies []Dependency `yaml:"dependencies"`
}

type Dependency struct {
	Name       string `yaml:"name"`
	Version    string `yaml:"version"`
	Repository string `yaml:"repository"`
}

func (g *clusterGitRepo) assembleChart(params *BaseParams) (*Chart, error) {
	templateRepo := g.templateRepo.GetLoc()
	return &Chart{
		APIVersion: "v2",
		Name:       params.Cluster,
		Version:    "1.0.0",
		Dependencies: []Dependency{
			{
				Repository: templateRepo,
				Name:       renameTemplateName(params.TemplateRelease.ChartName),
				Version:    params.TemplateRelease.ChartVersion,
			},
		},
	}, nil
}

// helm dependency 不支持 chart name 中包含 '.' 符号
func renameTemplateName(name string) string {
	templateName := []byte(name)
	for i := range templateName {
		if templateName[i] == '.' {
			templateName[i] = '_'
		}
	}
	return string(templateName)
}

func assembleRestart(templateName string) map[string]map[string]string {
	ret := make(map[string]map[string]string)
	ret[templateName] = make(map[string]string)
	ret[templateName]["restartTime"] = timeutil.Now(nil)
	return ret
}

func assembleTags(templateName string,
	tags []*tagmodels.Tag) map[string]map[string]map[string]string {
	const tagsKey = "tags"
	ret := make(map[string]map[string]map[string]string)
	ret[templateName] = make(map[string]map[string]string)
	ret[templateName][tagsKey] = make(map[string]string)
	for _, tag := range tags {
		ret[templateName][tagsKey][tag.Key] = tag.Value
	}
	return ret
}

func marshal(b *[]byte, err *error, data interface{}) {
	*b, *err = yaml.Marshal(data)
	if (*err) != nil {
		*err = perror.Wrap(herrors.ErrParamInvalid, (*err).Error())
	}
}
