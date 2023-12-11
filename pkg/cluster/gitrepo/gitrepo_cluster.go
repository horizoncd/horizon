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

package gitrepo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	gitlablib "github.com/horizoncd/horizon/lib/gitlab"
	"github.com/horizoncd/horizon/pkg/application/models"
	pkgcommon "github.com/horizoncd/horizon/pkg/common"
	"github.com/horizoncd/horizon/pkg/config/template"
	perror "github.com/horizoncd/horizon/pkg/errors"
	regionmodels "github.com/horizoncd/horizon/pkg/region/models"
	tagmodels "github.com/horizoncd/horizon/pkg/tag/models"
	trmodels "github.com/horizoncd/horizon/pkg/templaterelease/models"
	"github.com/horizoncd/horizon/pkg/templaterepo"
	"github.com/horizoncd/horizon/pkg/util/angular"
	utilcommon "github.com/horizoncd/horizon/pkg/util/common"
	"github.com/horizoncd/horizon/pkg/util/log"
	timeutil "github.com/horizoncd/horizon/pkg/util/time"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v3"
	kyaml "sigs.k8s.io/yaml"
)

const (
	GitOpsBranch        = "gitops"
	PipelineValueParent = "pipeline"
)

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

type CreateClusterParams struct {
	*BaseParams
	Tags  []*tagmodels.Tag
	Image string
}

type UpdateClusterParams struct {
	*BaseParams
}

type RepoInfo struct {
	GitRepoURL string
	ValueFiles []string
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

type ClusterTemplate struct {
	Name    string
	Release string
}

// Deprecated: for internal usage
type UpgradeValuesParam struct {
	Application   string
	Cluster       string
	Template      *ClusterTemplate
	TargetRelease *trmodels.TemplateRelease
	BuildConfig   *template.BuildConfig
}

type ReadFileParam struct {
	Bytes    []byte
	Err      error
	FileName string
}

// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/gitrepo/gitrepo_cluster_mock.go -package=mock_gitrepo
type ClusterGitRepo interface {
	GetCluster(ctx context.Context, application, cluster, templateName string) (*ClusterFiles, error)
	GetClusterValueFiles(ctx context.Context,
		application, cluster string) ([]ClusterValueFile, error)
	// GetClusterTemplate parses cluster's template name and release from GitopsFileChart
	GetClusterTemplate(ctx context.Context, application, cluster string) (*ClusterTemplate, error)
	CreateCluster(ctx context.Context, params *CreateClusterParams) error
	UpdateCluster(ctx context.Context, params *UpdateClusterParams) error
	DeleteCluster(ctx context.Context, application, cluster string, clusterID uint) error
	HardDeleteCluster(ctx context.Context, application, cluster string) error
	// CompareConfig compare config of `from` commit with `to` commit.
	// if `from` or `to` is nil, compare the master branch with gitops branch
	CompareConfig(ctx context.Context, application, cluster string, from, to *string) (string, error)
	// MergeBranch merge branch and return target branch's newest commit
	MergeBranch(ctx context.Context, application, cluster, sourceBranch,
		targetBranch string, pipelineRunID *uint) (_ string, err error)
	GetPipelineOutput(ctx context.Context, application, cluster string, template string) (interface{}, error)
	UpdatePipelineOutput(ctx context.Context, application, cluster, template string,
		pipelineOutput interface{}) (string, error)
	// UpdateRestartTime update restartTime in git repo for restart
	// TODO(gjq): some template cannot restart, for example serverless, how to do it ?
	GetRestartTime(ctx context.Context, application, cluster string,
		template string) (string, error)
	UpdateRestartTime(ctx context.Context, application, cluster, template string) (string, error)
	GetConfigCommit(ctx context.Context, application, cluster string) (*ClusterCommit, error)
	GetRepoInfo(ctx context.Context, application, cluster string) *RepoInfo
	GetEnvValue(ctx context.Context, application, cluster, templateName string) (*EnvValue, error)
	// Rollback rolls gitOps branch back to a specific commit if there are diffs
	Rollback(ctx context.Context, application, cluster, commit string) (string, error)
	UpdateTags(ctx context.Context, application, cluster, templateName string,
		tags []*tagmodels.Tag) error
	DefaultBranch() string
	// Deprecated: for internal usage, v1 to v2
	UpgradeCluster(ctx context.Context, param *UpgradeValuesParam) (string, error)
	// GetManifest returns manifest with specific revision, defaults to gitops branch
	GetManifest(ctx context.Context, application,
		cluster string, commit *string) (*pkgcommon.Manifest, error)
	// CheckAndSyncGitOpsBranch checks and sync if gitops branch is not up-to-date with master branch
	// for internal usage
	CheckAndSyncGitOpsBranch(ctx context.Context, application, cluster, commit string) error
	// SyncGitOpsBranch syncs gitops branch to up-to-date with master branch
	SyncGitOpsBranch(ctx context.Context, application, cluster string) error
}
type clusterGitopsRepo struct {
	gitlabLib              gitlablib.Interface
	clustersGroup          *gitlab.Group
	recyclingClustersGroup *gitlab.Group
	templateRepo           templaterepo.TemplateRepo
	defaultBranch          string
	defaultVisibility      string
}

func NewClusterGitlabRepo(ctx context.Context, rootGroup *gitlab.Group,
	templateRepo templaterepo.TemplateRepo,
	gitlabLib gitlablib.Interface, defaultBranch string, defaultVisibility string) (ClusterGitRepo, error) {
	clustersGroup, err := gitlabLib.GetCreatedGroup(ctx, rootGroup.ID, rootGroup.FullPath,
		common.GitopsGroupClusters, defaultVisibility)
	if err != nil {
		return nil, err
	}
	recyclingClustersGroup, err := gitlabLib.GetCreatedGroup(ctx,
		rootGroup.ID, rootGroup.FullPath, common.GitopsGroupRecyclingClusters, defaultVisibility)
	if err != nil {
		return nil, err
	}
	return &clusterGitopsRepo{
		gitlabLib:              gitlabLib,
		clustersGroup:          clustersGroup,
		recyclingClustersGroup: recyclingClustersGroup,
		templateRepo:           templateRepo,
		defaultBranch:          defaultBranch,
		defaultVisibility:      defaultVisibility,
	}, nil
}

func (g *clusterGitopsRepo) GetCluster(ctx context.Context,
	application, cluster, templateName string) (_ *ClusterFiles, err error) {
	const op = "cluster git repo: get cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get template and pipeline from gitlab
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	var applicationBytes, pipelineBytes, manifestBytes []byte
	var err1, err2, err3 error

	var wg sync.WaitGroup
	wg.Add(3)
	go func() {
		defer wg.Done()
		pipelineBytes, err1 = g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFilePipeline)
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
		applicationBytes, err2 = g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFileApplication)
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
		manifestBytes, err3 = g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFileManifest)
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
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
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
	var manifest = pkgcommon.Manifest{}
	if manifestBytes != nil {
		err = json.Unmarshal(manifestBytes, &manifest)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
	}

	pipelineJSONBlob, err := func() (map[string]interface{}, error) {
		if pipelineBytes == nil {
			return nil, nil
		}
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

	applicationJSONBlob, err := func() (map[string]interface{}, error) {
		if applicationBytes == nil {
			return nil, nil
		}
		jsonBlob, ok := applicationJSONBlobWithTemplate[templateName]
		if !ok {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"template value parent prefix not found, prefix = %s ", templateName)
		}
		return jsonBlob, nil
	}()
	if err != nil {
		return nil, err
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

func (g *clusterGitopsRepo) GetClusterValueFiles(ctx context.Context,
	application, cluster string) (_ []ClusterValueFile, err error) {
	const op = "cluster git repo: get cluster value files"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get  value file from git
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	cases := []ReadFileParam{
		{
			FileName: common.GitopsFileBase,
		}, {
			FileName: common.GitopsFileEnv,
		}, {
			FileName: common.GitopsFileApplication,
		}, {
			FileName: common.GitopsFileTags,
		}, {
			FileName: common.GitopsFileSRE,
		},
	}

	var wg sync.WaitGroup
	wg.Add(len(cases))
	for i := 0; i < len(cases); i++ {
		go func(index int) {
			defer wg.Done()
			cases[index].Bytes, cases[index].Err = g.gitlabLib.GetFile(ctx, pid,
				g.defaultBranch, cases[index].FileName)
			if cases[index].Err != nil {
				log.Warningf(ctx, "get file %s error, err = %s",
					cases[index].FileName, cases[index].Err.Error())
			}
		}(i)
	}
	wg.Wait()

	for i := 0; i < len(cases); i++ {
		if cases[i].Err != nil {
			if _, ok := perror.Cause(cases[i].Err).(*herrors.HorizonErrNotFound); !ok {
				log.Errorf(ctx, "get cluster value file error, err = %s", cases[i].Err.Error())
				return nil, cases[i].Err
			}
		}
	}

	// 2. check yaml format ok
	var clusterValueFiles []ClusterValueFile
	for _, oneCase := range cases {
		if oneCase.Err != nil {
			if _, ok := perror.Cause(oneCase.Err).(*herrors.HorizonErrNotFound); ok {
				continue
			}
		}
		var out map[interface{}]interface{}
		err = yaml.Unmarshal(oneCase.Bytes, &out)
		if err != nil {
			err = perror.Wrapf(herrors.ErrParamInvalid, "yaml Unmarshal err, file = %s", oneCase.FileName)
			break
		}

		clusterValueFiles = append(clusterValueFiles, ClusterValueFile{
			FileName: oneCase.FileName,
			Content:  out,
		})
	}
	if err != nil {
		return nil, err
	}

	return clusterValueFiles, nil
}

func (g *clusterGitopsRepo) GetClusterTemplate(ctx context.Context, application,
	cluster string) (*ClusterTemplate, error) {
	const op = "cluster git repo: get cluster template"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get Chart file from git
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	file, err := g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFileChart)
	if err != nil {
		return nil, err
	}
	var chart Chart
	err = yaml.Unmarshal(file, &chart)
	if err != nil {
		return nil, perror.Wrapf(herrors.ErrParamInvalid,
			"yaml Unmarshal err, file = %s", common.GitopsFileChart)
	}
	// extract release
	for _, dependency := range chart.Dependencies {
		if dependency.Name != "" && dependency.Version != "" {
			return &ClusterTemplate{
				Name:    dependency.Name,
				Release: parseReleaseName(dependency.Version),
			}, nil
		}
	}
	return nil, perror.Wrapf(herrors.ErrParamInvalid,
		"failed to get cluster template from chart")
}
func (g *clusterGitopsRepo) CreateCluster(ctx context.Context, params *CreateClusterParams) (err error) {
	const op = "cluster git repo: create cluster"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	// 1. create application group if necessary
	appGroup, err := g.gitlabLib.GetCreatedGroup(ctx, g.clustersGroup.ID,
		g.clustersGroup.FullPath, params.Application.Name, g.defaultVisibility)
	if err != nil {
		return err
	}

	// 3. create cluster repo under appGroup
	if _, err := g.gitlabLib.CreateProject(ctx, params.Cluster, appGroup.ID, g.defaultVisibility); err != nil {
		return err
	}

	// 3. create gitops branch from master
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, params.Application.Name, params.Cluster)
	if _, err := g.gitlabLib.CreateBranch(ctx, pid, GitOpsBranch, g.defaultBranch); err != nil {
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
		marshal(&manifestValueYAML, &err9, pkgcommon.Manifest{Version: params.BaseParams.Version})
	}

	chart, err := g.assembleChart(params.BaseParams)
	if err != nil {
		return err
	}
	marshal(&chartYAML, &err6, chart)

	for _, err := range []error{err1, err2, err3, err4, err5, err6, err7, err8, err9} {
		if err != nil {
			return err
		}
	}
	actions := func() []gitlablib.CommitAction {
		gitActions := []gitlablib.CommitAction{
			{
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileTags,
				Content:  string(tagsYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileBase,
				Content:  string(baseValueYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileEnv,
				Content:  string(envValueYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileSRE,
				Content:  string(sreValueYAML),
			}, {
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileChart,
				Content:  string(chartYAML),
			},
			// create GitopsFilePipelineOutput file first
			{
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFilePipelineOutput,
				Content:  "",
			},
			// create GitopsFileRestart file first
			{
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileRestart,
				Content:  string(restartYAML),
			},
		}

		if applicationYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileApplication,
				Content:  string(applicationYAML),
			})
		}
		if pipelineYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFilePipeline,
				Content:  string(pipelineYAML),
			})
		}
		if manifestValueYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileCreate,
				FilePath: common.GitopsFileManifest,
				Content:  string(manifestValueYAML),
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

	if _, err := g.gitlabLib.WriteFiles(ctx, pid, GitOpsBranch, commitMsg, nil, actions); err != nil {
		return err
	}

	return nil
}

func (g *clusterGitopsRepo) UpdateCluster(ctx context.Context, params *UpdateClusterParams) error {
	const op = "cluster git repo: update cluster"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	// 1. write files to repo
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, params.Application.Name, params.Cluster)
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

	actions, err := func() ([]gitlablib.CommitAction, error) {
		gitActions := []gitlablib.CommitAction{
			{
				Action:   gitlablib.FileUpdate,
				FilePath: common.GitopsFileBase,
				Content:  string(baseValueYAML),
			}, {
				Action:   gitlablib.FileUpdate,
				FilePath: common.GitopsFileChart,
				Content:  string(chartYAML),
			},
		}

		templateUpdate, pipelineUpdate, err := func() (gitlablib.FileAction, gitlablib.FileAction, error) {
			applicationUpdate, pipelineUpdate := gitlablib.FileCreate, gitlablib.FileCreate
			if applicationYAML != nil || pipelineYAML != nil {
				files, err := g.GetCluster(ctx, params.Application.Name, params.Cluster,
					params.TemplateRelease.TemplateName)
				if err != nil {
					return applicationUpdate, pipelineUpdate, err
				}
				if files.ApplicationJSONBlob != nil {
					applicationUpdate = gitlablib.FileUpdate
				}
				if files.PipelineJSONBlob != nil {
					pipelineUpdate = gitlablib.FileUpdate
				}
			}
			return applicationUpdate, pipelineUpdate, nil
		}()
		if err != nil {
			return nil, err
		}

		if applicationYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   templateUpdate,
				FilePath: common.GitopsFileApplication,
				Content:  string(applicationYAML),
			})
		}
		if pipelineYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   pipelineUpdate,
				FilePath: common.GitopsFilePipeline,
				Content:  string(pipelineYAML),
			})
		}
		if envValueYAML != nil {
			gitActions = append(gitActions, gitlablib.CommitAction{
				Action:   gitlablib.FileUpdate,
				FilePath: common.GitopsFileEnv,
				Content:  string(envValueYAML),
			})
		}
		return gitActions, nil
	}()
	if err != nil {
		return err
	}

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
	if _, err := g.gitlabLib.WriteFiles(ctx, pid, GitOpsBranch, commitMsg, nil, actions); err != nil {
		return err
	}

	return nil
}

func (g *clusterGitopsRepo) DeleteCluster(ctx context.Context,
	application, cluster string, clusterID uint) (err error) {
	const op = "cluster git repo: delete cluster"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. create application group if necessary
	_, err = g.gitlabLib.GetGroup(ctx, fmt.Sprintf("%v/%v", g.recyclingClustersGroup.FullPath, application))
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
		_, err = g.gitlabLib.CreateGroup(ctx, application, application,
			&g.recyclingClustersGroup.ID, g.defaultVisibility)
		if err != nil {
			return err
		}
	}

	// 1. delete gitlab project
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	// 1.1 edit project's name and path to {cluster}-{clusterID}
	newName := fmt.Sprintf("%v-%d", cluster, clusterID)
	newPath := newName
	if err := g.gitlabLib.EditNameAndPathForProject(ctx, pid, &newName, &newPath); err != nil {
		return err
	}

	// 1.2 transfer project to RecyclingParent
	newPid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, newPath)
	return g.gitlabLib.TransferProject(ctx, newPid,
		fmt.Sprintf("%v/%v", g.recyclingClustersGroup.FullPath, application))
}

func (g *clusterGitopsRepo) HardDeleteCluster(ctx context.Context, application,
	cluster string) (err error) {
	const op = "cluster git repo: hard delete cluster"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	return g.gitlabLib.DeleteProject(ctx, pid)
}

func (g *clusterGitopsRepo) CompareConfig(ctx context.Context, application,
	cluster string, from, to *string) (_ string, err error) {
	const op = "cluster git repo: compare config"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)

	var compare *gitlab.Compare
	if from == nil || to == nil {
		compare, err = g.gitlabLib.Compare(ctx, pid, g.defaultBranch, GitOpsBranch, nil)
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

func (g *clusterGitopsRepo) MergeBranch(ctx context.Context, application, cluster,
	sourceBranch, targetBranch string, pipelineRunID *uint) (_ string, err error) {
	removeSourceBranch := false
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)

	var title string
	if pipelineRunID != nil {
		title = fmt.Sprintf("git merge %v into %v, pipelineRunID = %d",
			sourceBranch, targetBranch, *pipelineRunID)
	} else {
		title = fmt.Sprintf("git merge %v into %v", sourceBranch, targetBranch)
	}

	var mr *gitlab.MergeRequest
	mrs, err := g.gitlabLib.ListMRs(ctx, pid, sourceBranch,
		targetBranch, common.GitopsMergeRequestStateOpen)
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
				len(mrs), sourceBranch, targetBranch)
			for i := 1; i < len(mrs); i++ {
				_, err := g.gitlabLib.CloseMR(ctx, pid, mrs[i].IID)
				if err != nil {
					return "", err
				}
			}
		}
	} else {
		// create new mr
		mr, err = g.gitlabLib.CreateMR(ctx, pid, sourceBranch, targetBranch, title)
		if err != nil {
			return "", perror.WithMessage(err, "failed to create new merge request")
		}
	}

	mr, err = g.gitlabLib.AcceptMR(ctx, pid, mr.IID, &title, &removeSourceBranch)
	if err != nil {
		return "", perror.WithMessage(err, "failed to accept merge request")
	}
	return mr.MergeCommitSHA, nil
}

func (g *clusterGitopsRepo) GetManifest(ctx context.Context, application,
	cluster string, commit *string) (*pkgcommon.Manifest, error) {
	const op = "cluster git repo: get manifest"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	var content []byte
	var err error
	if commit != nil {
		content, err = g.gitlabLib.GetFile(ctx, pid, *commit, common.GitopsFileManifest)
	} else {
		content, err = g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFileManifest)
	}
	if err != nil {
		return nil, err
	}
	manifest, err := func() (*pkgcommon.Manifest, error) {
		manifestOutputBytes, err := kyaml.YAMLToJSON(content)
		if err != nil {
			return nil, err
		}
		ret := pkgcommon.Manifest{}
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

func (g *clusterGitopsRepo) GetPipelineOutput(ctx context.Context, application, cluster string,
	template string) (interface{}, error) {
	ret := make(map[string]interface{})
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	content, err := g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFilePipelineOutput)
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
	return pipelineOutput, nil
}

func (g *clusterGitopsRepo) getPipelineOutput(ctx context.Context,
	application, cluster string) (map[string]map[string]interface{}, error) {
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	content, err := g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFilePipelineOutput)
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

func (g *clusterGitopsRepo) UpdatePipelineOutput(ctx context.Context, application, cluster, template string,
	pipelineOutput interface{}) (commitID string, err error) {
	const op = "cluster git repo: update pipeline output"
	defer wlog.Start(ctx, op).StopPrint()

	pipelineOutPutInternalFormat, err := func() (map[string]interface{}, error) {
		bytes, err := json.Marshal(pipelineOutput)
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

	pipelineOutPutValueParent := template
	newPipelineOutputContent := make(map[string]interface{})
	var PipelineOutPutFileExist bool
	_, err = g.getPipelineOutput(ctx, application, cluster)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return "", err
		}
		newPipelineOutputContent[pipelineOutPutValueParent] = pipelineOutPutInternalFormat
	} else {
		PipelineOutPutFileExist = true
		// if current exist, just override it
		newPipelineOutputContent[pipelineOutPutValueParent] = pipelineOutPutInternalFormat
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
			FilePath: common.GitopsFilePipelineOutput,
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
	}, pipelineOutput)

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	commit, err := g.gitlabLib.WriteFiles(ctx, pid, GitOpsBranch, commitMsg, nil, actions)
	if err != nil {
		return "", perror.WithMessage(err, "failed to write gitlab files")
	}
	return commit.ID, nil
}

func (g *clusterGitopsRepo) GetRestartTime(ctx context.Context, application, cluster string,
	template string) (string, error) {
	ret := make(map[string]map[string]string)
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	content, err := g.gitlabLib.GetFile(ctx, pid, g.defaultBranch, common.GitopsFileRestart)
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

func (g *clusterGitopsRepo) UpdateRestartTime(ctx context.Context,
	application, cluster, template string) (_ string, err error) {
	const op = "cluster git repo: update restartTime"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return "", err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)

	var restartYAML []byte
	var err1 error
	marshal(&restartYAML, &err1, assembleRestart(template))
	if err1 != nil {
		return "", err1
	}

	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: common.GitopsFileRestart,
			Content:  string(restartYAML),
		},
	}

	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "restart cluster",
		Cluster:  angular.StringPtr(cluster),
	}, nil)

	// update in defaultBranch directly
	commit, err := g.gitlabLib.WriteFiles(ctx, pid, g.defaultBranch, commitMsg, nil, actions)
	if err != nil {
		return "", err
	}

	return commit.ID, nil
}

func (g *clusterGitopsRepo) GetConfigCommit(ctx context.Context,
	application, cluster string) (_ *ClusterCommit, err error) {
	const op = "cluster git repo: get config commit"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)

	var branchMaster, branchGitops *gitlab.Branch
	var err1, err2 error
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		branchMaster, err1 = g.gitlabLib.GetBranch(ctx, pid, g.defaultBranch)
	}()
	go func() {
		defer wg.Done()
		branchGitops, err2 = g.gitlabLib.GetBranch(ctx, pid, GitOpsBranch)
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

func (g *clusterGitopsRepo) GetRepoInfo(ctx context.Context, application, cluster string) *RepoInfo {
	repoURL := g.gitlabLib.GetHTTPURL(ctx)
	return &RepoInfo{
		GitRepoURL: fmt.Sprintf("%v/%v/%v/%v.git", repoURL, g.clustersGroup.FullPath, application, cluster),
		ValueFiles: []string{common.GitopsFileApplication, common.GitopsFilePipelineOutput,
			common.GitopsFileEnv, common.GitopsFileBase, common.GitopsFileTags, common.GitopsFileRestart, common.GitopsFileSRE},
	}
}

func (g *clusterGitopsRepo) GetEnvValue(ctx context.Context,
	application, cluster, templateName string) (_ *EnvValue, err error) {
	const op = "cluster git repo: get config commit"
	defer wlog.Start(ctx, op).StopPrint()

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)

	bytes, err := g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, common.GitopsFileEnv)
	if err != nil {
		return nil, err
	}

	var envMap map[string]map[string]*EnvValue

	if err := yaml.Unmarshal(bytes, &envMap); err != nil {
		return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	return envMap[templateName][common.GitopsEnvValueNamespace], nil
}

func (g *clusterGitopsRepo) CheckAndSyncGitOpsBranch(ctx context.Context, application, cluster, commit string) error {
	changed, err := g.manifestVersionChanged(ctx, application, cluster, commit)
	if err != nil {
		return err
	}
	if changed {
		err = g.SyncGitOpsBranch(ctx, application, cluster)
		if err != nil {
			return err
		}
	}
	return nil
}

func (g *clusterGitopsRepo) Rollback(ctx context.Context, application, cluster, commit string) (_ string, err error) {
	const op = "cluster git repo: rollback"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return "", err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)

	// compare commit straight diffs
	compare, err := g.gitlabLib.Compare(ctx, pid, GitOpsBranch, commit, utilcommon.BoolPtr(true))
	if err != nil {
		return "", err
	}
	if compare.Diffs == nil {
		return "", perror.Wrapf(herrors.ErrParamInvalid,
			"dose not support empty rollback, rollback commit = %s", commit)
	}

	type actionCase struct {
		action *gitlablib.CommitAction
		err    error
	}
	cases := make([]actionCase, len(compare.Diffs))
	var wg sync.WaitGroup
	for i := range compare.Diffs {
		i := i
		wg.Add(1)
		// generate a commit action for rollback based on diff
		go func() {
			defer wg.Done()
			action, err := g.revertAction(ctx, application, cluster, commit, *compare.Diffs[i])
			cases[i] = actionCase{
				action: action,
				err:    err,
			}
		}()
	}
	wg.Wait()

	var actions []gitlablib.CommitAction
	for _, oneCase := range cases {
		if oneCase.err != nil {
			return "", oneCase.err
		}
		actions = append(actions, *oneCase.action)
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

	newCommit, err := g.gitlabLib.WriteFiles(ctx, pid, GitOpsBranch, commitMsg, nil, actions)
	if err != nil {
		return "", err
	}

	return newCommit.ID, nil
}

func (g *clusterGitopsRepo) UpdateTags(ctx context.Context, application, cluster, templateName string,
	tags []*tagmodels.Tag) (err error) {
	const op = "cluster git repo: update tags"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)

	var tagsYAML []byte
	marshal(&tagsYAML, &err, assembleTags(templateName, tags))
	if err != nil {
		return err
	}

	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: common.GitopsFileTags,
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

	_, err = g.gitlabLib.WriteFiles(ctx, pid, GitOpsBranch, commitMsg, nil, actions)
	if err != nil {
		return err
	}

	return nil
}

func (g *clusterGitopsRepo) DefaultBranch() string {
	return g.defaultBranch
}

func (g *clusterGitopsRepo) UpgradeCluster(ctx context.Context,
	param *UpgradeValuesParam) (string, error) {
	const op = "cluster git repo: upgrade cluster"
	defer wlog.Start(ctx, op).StopPrint()
	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return "", err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath,
		param.Application, param.Cluster)

	type upgradeValueBytes struct {
		fileName      string
		sourceBytes   []byte
		upgradedBytes []byte
		err           error
	}

	cases := []*upgradeValueBytes{
		{
			fileName: common.GitopsFileBase,
		}, {
			fileName: common.GitopsFileEnv,
		}, {
			fileName: common.GitopsFileApplication,
		}, {
			fileName: common.GitopsFileTags,
		}, {
			fileName: common.GitopsFileSRE,
		}, {
			fileName: common.GitopsFilePipeline,
		}, {
			fileName: common.GitopsFilePipelineOutput,
		}, {
			fileName: common.GitopsFileChart,
		}, {
			fileName: common.GitopsFileRestart,
		}, {
			fileName: common.GitopsFileManifest,
		},
	}

	// 1. read files
	var wgReadFile sync.WaitGroup
	wgReadFile.Add(len(cases))
	for i := range cases {
		i := i
		go func() {
			defer wgReadFile.Done()
			cases[i].sourceBytes, cases[i].err = g.readFile(ctx, param.Application,
				param.Cluster, cases[i].fileName, nil)
		}()
	}
	wgReadFile.Wait()
	for _, oneCase := range cases {
		if oneCase.fileName != common.GitopsFileManifest {
			if oneCase.err != nil {
				log.Errorf(ctx, "get cluster value file error, file = %s, err = %s",
					oneCase.fileName, oneCase.err.Error())
				return "", oneCase.err
			}
		} else {
			if oneCase.err != nil {
				if _, ok := perror.Cause(oneCase.err).(*herrors.HorizonErrNotFound); !ok {
					log.Errorf(ctx, "get cluster value file error, file = %s, err = %s",
						oneCase.fileName, oneCase.err.Error())
					return "", oneCase.err
				}
			}
		}
	}

	updateChartValue := func(sourceBytes []byte) ([]byte, error) {
		var (
			upgradedBytes []byte
			err           error
			chart         Chart
		)
		err = yaml.Unmarshal(sourceBytes, &chart)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"yaml Unmarshal err, file = %s, err = %s", common.GitopsFileChart, err.Error())
		}
		// update template
		if len(chart.Dependencies) > 0 {
			dep := chart.Dependencies[0]
			chart.Dependencies[0] = Dependency{
				Name:       param.TargetRelease.ChartName,
				Version:    param.TargetRelease.ChartVersion,
				Repository: dep.Repository,
			}
		}
		marshal(&upgradedBytes, &err, &chart)
		return upgradedBytes, err
	}
	updateBaseValue := func(sourceBytes []byte) ([]byte, error) {
		var (
			upgradedBytes []byte
			err           error
		)
		valueMap := make(map[string]map[string]*BaseValue)
		err = yaml.Unmarshal(sourceBytes, &valueMap)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"yaml Unmarshal err, file = %s, err = %s", common.GitopsFileBase, err.Error())
		}
		baseValue, ok := valueMap[param.Template.Name][common.GitopsBaseValueNamespace]
		if !ok {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"value parent err, file = %s, parent = %s", common.GitopsFileBase, param.Template.Name)
		}
		baseValue.Template = &BaseValueTemplate{
			Name:    param.TargetRelease.TemplateName,
			Release: param.TargetRelease.ChartVersion,
		}
		retMap := make(map[string]map[string]*BaseValue)
		retMap[param.TargetRelease.TemplateName] = map[string]*BaseValue{
			common.GitopsBaseValueNamespace: baseValue,
		}
		marshal(&upgradedBytes, &err, &retMap)
		return upgradedBytes, err
	}
	assembleManifest := func() ([]byte, error) {
		var (
			upgradedBytes []byte
			err           error
		)
		marshal(&upgradedBytes, &err, &pkgcommon.Manifest{Version: common.MetaVersion2})
		if err != nil {
			log.Errorf(ctx, "marshal manifest error, err = %s", err.Error())
		}
		return upgradedBytes, nil
	}
	updatePipelineValue := func(sourceBytes []byte) ([]byte, error) {
		if sourceBytes == nil {
			return nil, nil
		}

		var (
			inMap         map[string]map[string]interface{}
			upgradedBytes []byte
			err           error
		)
		err = yaml.Unmarshal(sourceBytes, &inMap)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"yaml Unmarshal err, file = %s, err = %s", common.GitopsFilePipeline, err.Error())
		}
		antScript, ok := inMap[param.Template.Name]["buildxml"]
		if !ok {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"value parent err, file = %s, parent = %s",
				common.GitopsFilePipeline, param.Template.Name)
		}
		retMap := map[string]map[string]interface{}{
			PipelineValueParent: {
				"buildInfo": map[string]interface{}{
					"buildTool": "ant",
					"buildxml":  antScript,
				},
				"buildType":   "netease-normal",
				"environment": param.BuildConfig.Environment,
				"language":    param.BuildConfig.Language,
			},
		}
		marshal(&upgradedBytes, &err, &retMap)
		return upgradedBytes, err
	}
	updateApplicationValue := func(sourceBytes []byte) ([]byte, error) {
		if sourceBytes == nil {
			return nil, nil
		}

		var (
			inMap         map[string]map[string]interface{}
			upgradedBytes []byte
			err           error
		)
		err = yaml.Unmarshal(sourceBytes, &inMap)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"yaml Unmarshal err, file = %s, err = %s", common.GitopsFileApplication, err.Error())
		}
		midMap, ok := inMap[param.Template.Name]
		if !ok {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"value parent err, file = %s, parent = %s",
				common.GitopsFileApplication, param.Template.Name)
		}
		// convert params to envs
		func() {
			appMap, ok := midMap["app"].(map[string]interface{})
			if !ok {
				return
			}
			paramsMap, ok := appMap["params"].(map[string]interface{})
			if ok {
				var envsArray []map[string]interface{}
				for k, v := range paramsMap {
					envsArray = append(envsArray, map[string]interface{}{
						"name":  k,
						"value": v,
					})
				}
				delete(appMap, "params")
				appMap["envs"] = envsArray
			}
		}()
		retMap := make(map[string]map[string]interface{})
		retMap[param.TargetRelease.TemplateName] = midMap
		marshal(&upgradedBytes, &err, &retMap)
		return upgradedBytes, err
	}
	updateValueParent := func(fileName string, sourceBytes []byte) ([]byte, error) {
		if sourceBytes == nil {
			return nil, nil
		}

		var (
			inMap         map[string]interface{}
			upgradedBytes []byte
			err           error
		)
		err = yaml.Unmarshal(sourceBytes, &inMap)
		if err != nil {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"yaml Unmarshal err, file = %s, err = %s", fileName, err.Error())
		}
		valueMap, ok := inMap[param.Template.Name]
		if !ok {
			return nil, perror.Wrapf(herrors.ErrParamInvalid,
				"value parent err, file = %s, parent = %s", fileName, param.Template.Name)
		}
		retMap := map[string]interface{}{
			param.TargetRelease.TemplateName: valueMap,
		}
		marshal(&upgradedBytes, &err, &retMap)
		return upgradedBytes, err
	}

	// 2. upgrade value bytes
	var wgUpdateValue sync.WaitGroup
	wgUpdateValue.Add(len(cases))
	for i := range cases {
		i := i
		switch cases[i].fileName {
		case common.GitopsFileChart:
			go func() {
				defer wgUpdateValue.Done()
				cases[i].upgradedBytes, cases[i].err = updateChartValue(cases[i].sourceBytes)
			}()
		case common.GitopsFileBase:
			go func() {
				defer wgUpdateValue.Done()
				cases[i].upgradedBytes, cases[i].err = updateBaseValue(cases[i].sourceBytes)
			}()
		case common.GitopsFileManifest:
			go func() {
				defer wgUpdateValue.Done()
				cases[i].upgradedBytes, _ = assembleManifest()
			}()
		case common.GitopsFilePipeline:
			go func() {
				defer wgUpdateValue.Done()
				cases[i].upgradedBytes, cases[i].err = updatePipelineValue(cases[i].sourceBytes)
			}()
		case common.GitopsFileApplication:
			go func() {
				defer wgUpdateValue.Done()
				cases[i].upgradedBytes, cases[i].err = updateApplicationValue(cases[i].sourceBytes)
			}()
		default:
			// update value parent
			go func() {
				defer wgUpdateValue.Done()
				cases[i].upgradedBytes, cases[i].err =
					updateValueParent(cases[i].fileName, cases[i].sourceBytes)
			}()
		}
	}
	wgUpdateValue.Wait()

	// 3. write files
	var gitActions []gitlablib.CommitAction
	for _, oneCase := range cases {
		if oneCase.fileName != common.GitopsFileManifest {
			if oneCase.err != nil {
				return "", oneCase.err
			}
			if oneCase.sourceBytes != nil {
				gitActions = append(gitActions, gitlablib.CommitAction{
					Action:   gitlablib.FileUpdate,
					FilePath: oneCase.fileName,
					Content:  string(oneCase.upgradedBytes),
				})
			}
		} else {
			if oneCase.err != nil {
				gitActions = append(gitActions, gitlablib.CommitAction{
					Action:   gitlablib.FileCreate,
					FilePath: oneCase.fileName,
					Content:  string(oneCase.upgradedBytes),
				})
			} else {
				gitActions = append(gitActions, gitlablib.CommitAction{
					Action:   gitlablib.FileUpdate,
					FilePath: oneCase.fileName,
					Content:  string(oneCase.upgradedBytes),
				})
			}
		}
	}
	commitMsg := angular.CommitMessage("cluster", angular.Subject{
		Operator: currentUser.GetName(),
		Action:   "upgrade cluster",
		Cluster:  angular.StringPtr(param.Cluster),
	}, struct {
		SourceTemplate ClusterTemplate `json:"sourceTemplate"`
		TargetTemplate ClusterTemplate `json:"targetTemplate"`
	}{
		SourceTemplate: *param.Template,
		TargetTemplate: ClusterTemplate{
			Name:    param.TargetRelease.TemplateName,
			Release: param.TargetRelease.Name,
		},
	})
	newCommit, err := g.gitlabLib.WriteFiles(ctx, pid, GitOpsBranch, commitMsg, nil, gitActions)
	if err != nil {
		return "", err
	}
	return newCommit.ID, nil
}

// assembleApplicationValue assemble application.yaml data
func (g *clusterGitopsRepo) assembleApplicationValue(params *BaseParams) map[string]map[string]interface{} {
	ret := make(map[string]map[string]interface{})
	ret[params.TemplateRelease.ChartName] = params.ApplicationJSONBlob
	return ret
}

// assembleApplicationValue assemble pipeline.yaml data
func (g *clusterGitopsRepo) assemblePipelineValue(params *BaseParams) map[string]map[string]interface{} {
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
func (g *clusterGitopsRepo) assembleSREValue(params *CreateClusterParams) map[string]interface{} {
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

func (g *clusterGitopsRepo) assembleEnvValue(params *BaseParams) map[string]map[string]*EnvValue {
	envMap := make(map[string]*EnvValue)
	envMap[common.GitopsEnvValueNamespace] = &EnvValue{
		Environment: params.Environment,
		Region:      params.RegionEntity.Name,
		Namespace:   getNamespace(params),
		BaseRegistry: strings.TrimPrefix(strings.TrimPrefix(
			params.RegionEntity.Registry.Server, "https://"), "http://"),
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
func (g *clusterGitopsRepo) assembleBaseValue(params *BaseParams) map[string]map[string]*BaseValue {
	baseMap := make(map[string]*BaseValue)
	baseMap[common.GitopsBaseValueNamespace] = &BaseValue{
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

func (g *clusterGitopsRepo) assembleChart(params *BaseParams) (*Chart, error) {
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

// revertAction generates a commit action based on the diff to simulate 'git revert'
// ref: https://git-scm.com/docs/git-revert
// an example of all types of file changes:
//
//	"diffs": [
//		{
//			"old_path": "Chart_bak.yaml",
//			"new_path": "Chart.yaml",
//			"new_file": false,
//			"renamed_file": true,  // Chart.yaml is renamed to Chart_bak.yaml in gitOps branch
//			"deleted_file": false,
//			"diff": ""
//		}, {
//			"old_path": "README.md",
//			"new_path": "README.md",
//			"new_file": true, // README.md is deleted in gitOps branch
//			"renamed_file": false,
//			"deleted_file": false,
//			"diff": "..."
//		}, {
//			"old_path": "application.yaml", // application.yaml is updated in gitOps branch
//			"new_path": "application.yaml",
//			"new_file": false,
//			"renamed_file": false,
//			"deleted_file": false,
//			"diff": "..."
//		}, {
//			"old_path": "new_file.yaml",
//			"new_path": "new_file.yaml",
//			"new_file": false,
//			"renamed_file": false,
//			"deleted_file": true, // new_file.yaml is created in gitOps branch
//			"diff": "..."
//		}
//	]
func (g *clusterGitopsRepo) revertAction(ctx context.Context, application, cluster,
	commit string, diff gitlab.Diff) (*gitlablib.CommitAction, error) {
	if diff.DeletedFile {
		// file is deleted from gitops branch to the commit
		return &gitlablib.CommitAction{
			Action:   gitlablib.FileDelete,
			FilePath: diff.OldPath,
		}, nil
	}
	if diff.NewFile {
		// file is created from gitops branch to the commit
		file, err := g.readFile(ctx, application, cluster, diff.NewPath, &commit)
		if err != nil {
			return nil, err
		}
		return &gitlablib.CommitAction{
			Action:   gitlablib.FileCreate,
			FilePath: diff.NewPath,
			Content:  string(file),
		}, nil
	}
	if diff.RenamedFile {
		// file is renamed from gitops branch to the commit
		return &gitlablib.CommitAction{
			Action:       gitlablib.FileMove,
			FilePath:     diff.NewPath,
			PreviousPath: diff.OldPath,
		}, nil
	}
	// file is updated from gitops branch to the commit if flags are all false
	file, err := g.readFile(ctx, application, cluster, diff.NewPath, &commit)
	if err != nil {
		return nil, err
	}
	return &gitlablib.CommitAction{
		Action:   gitlablib.FileUpdate,
		FilePath: diff.NewPath,
		Content:  string(file),
	}, nil
}

// readFile gets file for specific revision, defaults to gitOps branch
func (g *clusterGitopsRepo) readFile(ctx context.Context, application, cluster,
	fileName string, commit *string) ([]byte, error) {
	pid := fmt.Sprintf("%v/%v/%v", g.clustersGroup.FullPath, application, cluster)
	if commit != nil {
		return g.gitlabLib.GetFile(ctx, pid, *commit, fileName)
	}
	return g.gitlabLib.GetFile(ctx, pid, GitOpsBranch, fileName)
}

// for internal usage
// manifestVersionChanged determines whether manifest version is changed
func (g *clusterGitopsRepo) manifestVersionChanged(ctx context.Context, application,
	cluster string, commit string) (bool, error) {
	currentManifest, err1 := g.GetManifest(ctx, application, cluster, nil)
	if err1 != nil {
		if _, ok := perror.Cause(err1).(*herrors.HorizonErrNotFound); !ok {
			log.Errorf(ctx, "get cluster manifest error, err = %s", err1.Error())
			return false, err1
		}
	}
	targetManifest, err2 := g.GetManifest(ctx, application, cluster, &commit)
	if err2 != nil {
		if _, ok := perror.Cause(err2).(*herrors.HorizonErrNotFound); !ok {
			log.Errorf(ctx, "get cluster manifest error, err = %s", err2.Error())
			return false, err2
		}
	}
	if err1 != nil && err2 != nil {
		// manifest does not exist in both revisions
		return false, nil
	}
	if err1 != nil || err2 != nil {
		// One exists and the other does not exist in two revisions
		return true, nil
	}
	return currentManifest.Version != targetManifest.Version, nil
}

// SyncGitOpsBranch for internal usage, syncs gitOps branch with default branch to avoid merge conflicts.
// Restart updates time in restart.yaml in default branch. When other actions update
// template prefix in gitOps branch, there are merge conflicts in restart.yaml because
// usual context lines of 'git diff' are three. Ref: https://git-scm.com/docs/git-diff
// For example:
//
//	<<<<<<< HEAD
//	javaapp:
//	  restartTime: "2025-02-19 10:24:52"
//	=======
//	rollout:
//	  restartTime: "2025-02-14 12:12:07"
//	>>>>>>> gitops
func (g *clusterGitopsRepo) SyncGitOpsBranch(ctx context.Context, application, cluster string) error {
	gitOpsBranch := GitOpsBranch
	defaultBranch := g.DefaultBranch()
	diff, err := g.CompareConfig(ctx, application, cluster,
		&gitOpsBranch, &defaultBranch)
	if err != nil {
		return err
	}
	if diff != "" {
		_, err = g.MergeBranch(ctx, application,
			cluster, defaultBranch, gitOpsBranch, nil)
		if err != nil {
			return err
		}
	}
	return nil
}

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
	ret := make(map[string]map[string]map[string]string)
	ret[templateName] = make(map[string]map[string]string)
	ret[templateName][common.GitopsKeyTags] = make(map[string]string)
	for _, tag := range tags {
		ret[templateName][common.GitopsKeyTags][tag.Key] = tag.Value
	}
	return ret
}

func marshal(b *[]byte, err *error, data interface{}) {
	buf := bytes.NewBuffer(make([]byte, 0))
	encoder := yaml.NewEncoder(buf)
	encoder.SetIndent(2)
	*err = encoder.Encode(data)
	if (*err) != nil {
		*err = perror.Wrap(herrors.ErrParamInvalid, (*err).Error())
	} else {
		*b = buf.Bytes()
	}
}

// parseReleaseName extracts release name from chart version
// such as:
//
//	v1.3.2-e0dee4e9 => v1.3.2
//	v1.2.6-ad3ac3cb700786fbb368988510b46b356a76c917 => v1.2.6
//	v1.0.0-rc1 => v1.0.0-rc1
func parseReleaseName(chartVersion string) string {
	pattern := regexp.MustCompile("(-([a-z0-9]){8}$)|(-([a-z0-9]){40}$)")
	b := []byte(chartVersion)
	loc := pattern.FindIndex(b)
	if len(loc) == 0 {
		return chartVersion
	}
	return string(b[:loc[0]])
}
