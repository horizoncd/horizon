package gitrepo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	"g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cluster/cd/argocd"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
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
                              ├── sre                         -- sre目录
                              │     └── sre.yaml              -- sre values数据
                              ├── system
                              │     ├── horizon.yaml          -- 基础数据
                              │     └── env.yaml              -- 环境相关数据
                              ├── pipeline
                              │     ├── pipeline.yaml         -- pipeline模板参数
                              │     └── pipeline-output.yaml  -- pipeline输出
							  └── argo
							        └── argo-application.yaml -- argo application
*/

const (
	_gitlabName = "compute"

	// _branchMaster is the main branch
	_branchMaster = "master"
	// _branchGitops is the gitops branch, values updated in this branch, then merge into the _branchMaster
	_branchGitops = "gitops"

	_filePathChart       = "Chart.yaml"
	_filePathApplication = "application.yaml"
	_filePathSRE         = "sre/sre.yaml"
	_filePathBase        = "system/horizon.yaml"
	_filePathEnv         = "system/env.yaml"
	_filePathPipeline    = "pipeline/pipeline.yaml"
	// _filePathPipelineOutput  = "pipeline/pipeline-output.yaml"
	_filePathArgoApplication = "argo/argo-application.yaml"
)

type Params struct {
	Cluster             string
	HelmRepoURL         string
	Environment         string
	RegionEntity        *regionmodels.RegionEntity
	PipelineJSONBlob    map[string]interface{}
	ApplicationJSONBlob map[string]interface{}
	TemplateRelease     *trmodels.TemplateRelease
	Application         *models.Application
}

type ClusterGitRepo interface {
	CreateCluster(ctx context.Context, params *Params) error
	UpdateCluster(ctx context.Context, params *Params) error
	DeleteCluster(ctx context.Context, application, cluster string, clusterID uint) error
	CompareConfig(ctx context.Context, application, cluster string) (string, error)
	MergeBranch(ctx context.Context, application, cluster string) error
	ReadArgoApplication(ctx context.Context, application, cluster string) ([]byte, error)
}

type clusterGitRepo struct {
	gitlabLib       gitlablib.Interface
	clusterRepoConf *gitlabconf.Repo
}

func NewClusterGitlabRepo(ctx context.Context, gitlabRepoConfig gitlabconf.RepoConfig,
	gitlabFactory gitlabfty.Factory) (ClusterGitRepo, error) {
	gitlabLib, err := gitlabFactory.GetByName(ctx, _gitlabName)
	if err != nil {
		return nil, err
	}
	return &clusterGitRepo{
		gitlabLib:       gitlabLib,
		clusterRepoConf: gitlabRepoConfig.Cluster,
	}, nil
}

func (g *clusterGitRepo) CreateCluster(ctx context.Context, params *Params) (err error) {
	const op = "cluster git repo: create cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. create application group if necessary
	var appGroup *gitlab.Group
	appGroup, err = g.gitlabLib.GetGroup(ctx, fmt.Sprintf("%v/%v", g.clusterRepoConf.Parent.Path, params.Application.Name))
	if err != nil {
		if errors.Status(err) != http.StatusNotFound {
			return errors.E(op, err)
		}
		appGroup, err = g.gitlabLib.CreateGroup(ctx, params.Application.Name,
			params.Application.Name, &g.clusterRepoConf.Parent.ID)
		if err != nil {
			return errors.E(op, err)
		}
	}

	// 2. create cluster repo under appGroup
	if _, err := g.gitlabLib.CreateProject(ctx, params.Cluster, appGroup.ID); err != nil {
		return errors.E(op, err)
	}

	// 3. write files to repo
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, params.Application.Name, params.Cluster)

	var applicationYAML, pipelineYAML, baseValueYAML, envValueYAML, sreValueYAML []byte
	var chartYAML, argoApplicationYAML []byte
	var err1, err2, err3, err4, err5, err6, err7 error
	marshal := func(b *[]byte, err *error, data interface{}) {
		*b, *err = yaml.Marshal(data)
	}
	marshal(&applicationYAML, &err1, g.assembleApplicationValue(params))
	marshal(&pipelineYAML, &err2, g.assemblePipelineValue(params))
	marshal(&baseValueYAML, &err3, g.assembleBaseValue(params))
	marshal(&envValueYAML, &err4, g.assembleEnvValue(params))
	marshal(&sreValueYAML, &err5, g.assembleSREValue(params))
	marshal(&chartYAML, &err6, g.assembleChart(params))
	marshal(&argoApplicationYAML, &err7, g.assembleArgoApplication(ctx, params))

	for _, err := range []error{err1, err2, err3, err4, err5, err6, err7} {
		if err != nil {
			return errors.E(op, err)
		}
	}

	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileCreate,
			FilePath: _filePathApplication,
			Content:  string(applicationYAML),
		}, {
			Action:   gitlablib.FileCreate,
			FilePath: _filePathPipeline,
			Content:  string(pipelineYAML),
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
		}, {
			Action:   gitlablib.FileCreate,
			FilePath: _filePathArgoApplication,
			Content:  string(argoApplicationYAML),
		},
	}

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

	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, err)
	}

	// 3. create gitops branch from master
	if _, err := g.gitlabLib.CreateBranch(ctx, pid, _branchGitops, _branchMaster); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *clusterGitRepo) UpdateCluster(ctx context.Context, params *Params) (err error) {
	const op = "cluster git repo: update cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. write files to repo
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, params.Application.Name, params.Cluster)
	var applicationYAML, pipelineYAML, baseValueYAML []byte
	var err1, err2, err3 error
	marshal := func(b *[]byte, err *error, data interface{}) {
		*b, *err = yaml.Marshal(data)
	}
	marshal(&applicationYAML, &err1, g.assembleApplicationValue(params))
	marshal(&pipelineYAML, &err2, g.assemblePipelineValue(params))
	marshal(&baseValueYAML, &err3, g.assembleBaseValue(params))

	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			return err
		}
	}

	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathApplication,
			Content:  string(applicationYAML),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathPipeline,
			Content:  string(pipelineYAML),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathBase,
			Content:  string(baseValueYAML),
		},
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

	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchGitops, commitMsg, nil, actions); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *clusterGitRepo) DeleteCluster(ctx context.Context, application, cluster string, clusterID uint) (err error) {
	const op = "cluster git repo: delete cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. create application group if necessary
	_, err = g.gitlabLib.GetGroup(ctx, fmt.Sprintf("%v/%v", g.clusterRepoConf.RecyclingParent.Path, application))
	if err != nil {
		if errors.Status(err) != http.StatusNotFound {
			return errors.E(op, err)
		}
		_, err = g.gitlabLib.CreateGroup(ctx, application, application, &g.clusterRepoConf.RecyclingParent.ID)
		if err != nil {
			return errors.E(op, err)
		}
	}

	// 1. delete gitlab project
	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)
	// 1.1 edit project's name and path to {cluster}-{clusterID}
	newName := fmt.Sprintf("%v-%d", cluster, clusterID)
	newPath := newName
	if err := g.gitlabLib.EditNameAndPathForProject(ctx, pid, &newName, &newPath); err != nil {
		return errors.E(op, err)
	}

	// 1.2 transfer project to RecyclingParent
	newPid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, newPath)
	if err := g.gitlabLib.TransferProject(ctx, newPid,
		fmt.Sprintf("%v/%v", g.clusterRepoConf.RecyclingParent.Path, application)); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *clusterGitRepo) CompareConfig(ctx context.Context, application, cluster string) (_ string, err error) {
	const op = "cluster git repo: compare config"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	compare, err := g.gitlabLib.Compare(ctx, pid, _branchMaster, _branchGitops, nil)
	if err != nil {
		return "", errors.E(op, err)
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

func (g *clusterGitRepo) MergeBranch(ctx context.Context, application, cluster string) (err error) {
	const op = "cluster git repo: merge branch"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	title := fmt.Sprintf("git merge %v", _branchGitops)
	mr, err := g.gitlabLib.CreateMR(ctx, pid, _branchGitops, _branchMaster, title)
	if err != nil {
		return errors.E(op, err)
	}

	// TODO(gjq) add this msg
	// mergeCommitMsg := ""
	removeSourceBranch := false
	if _, err := g.gitlabLib.AcceptMR(ctx, pid, mr.IID, nil, &removeSourceBranch); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *clusterGitRepo) ReadArgoApplication(ctx context.Context, application, cluster string) (_ []byte, err error) {
	const op = "cluster git repo: read argo application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pid := fmt.Sprintf("%v/%v/%v", g.clusterRepoConf.Parent.Path, application, cluster)

	bytes, err := g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathArgoApplication)
	if err != nil {
		return nil, errors.E(op, err)
	}

	return bytes, nil
}

// assembleApplicationValue assemble application.yaml data
func (g *clusterGitRepo) assembleApplicationValue(params *Params) map[string]map[string]interface{} {
	ret := make(map[string]map[string]interface{})
	ret[params.TemplateRelease.TemplateName] = params.ApplicationJSONBlob
	return ret
}

// assembleApplicationValue assemble pipeline.yaml data
func (g *clusterGitRepo) assemblePipelineValue(params *Params) map[string]map[string]interface{} {
	ret := make(map[string]map[string]interface{})
	ret[params.TemplateRelease.TemplateName] = params.PipelineJSONBlob
	return ret
}

// assembleSreValue assemble sre value data
func (g *clusterGitRepo) assembleSREValue(params *Params) map[string]interface{} {
	ret := make(map[string]interface{})
	ret[params.TemplateRelease.TemplateName] = make(map[string]string)
	return ret
}

type EnvValue struct {
	Environment  string `yaml:"environment"`
	Region       string `yaml:"region"`
	Namespace    string `yaml:"namespace"`
	BaseRegistry string `yaml:"baseRegistry"`
}

func (g *clusterGitRepo) assembleEnvValue(params *Params) map[string]map[string]*EnvValue {
	const envValueNamespace = "env"
	var namespace = fmt.Sprintf("%v-%v", params.Environment, params.Application.GroupID)
	envMap := make(map[string]*EnvValue)
	envMap[envValueNamespace] = &EnvValue{
		Environment: params.Environment,
		Region:      params.RegionEntity.Name,
		Namespace:   namespace,
		BaseRegistry: strings.TrimPrefix(strings.TrimPrefix(
			params.RegionEntity.Harbor.Server, "https://"), "http://"),
	}

	ret := make(map[string]map[string]*EnvValue)
	ret[params.TemplateRelease.TemplateName] = envMap
	return ret
}

type BaseValue struct {
	Application string             `yaml:"application"`
	Cluster     string             `yaml:"cluster"`
	Template    *BaseValueTemplate `yaml:"template"`
	Priority    string             `yaml:"priority"`
}

type BaseValueTemplate struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

// assembleBaseValue assemble base value. return a map, key is template name,
// and value is a map which key is "horizon", and value is *BaseValue
func (g *clusterGitRepo) assembleBaseValue(params *Params) map[string]map[string]*BaseValue {
	const baseValueNamespace = "horizon"
	baseMap := make(map[string]*BaseValue)
	baseMap[baseValueNamespace] = &BaseValue{
		Application: params.Application.Name,
		Cluster:     params.Cluster,
		Template: &BaseValueTemplate{
			Name:    params.TemplateRelease.TemplateName,
			Release: params.TemplateRelease.Name,
		},
		Priority: string(params.Application.Priority),
	}

	ret := make(map[string]map[string]*BaseValue)
	ret[params.TemplateRelease.TemplateName] = baseMap
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

func (g *clusterGitRepo) assembleChart(params *Params) *Chart {
	return &Chart{
		APIVersion: "v2",
		Name:       params.Cluster,
		Version:    "1.0.0",
		Dependencies: []Dependency{
			{
				Repository: params.HelmRepoURL,
				Name:       renameTemplateName(params.TemplateRelease.TemplateName),
				Version:    params.TemplateRelease.Name,
			},
		},
	}
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

func (g *clusterGitRepo) assembleArgoApplication(ctx context.Context, params *Params) *argocd.Application {
	const argoCDNamespace = "argocd"
	const finalizer = "resources-finalizer.argocd.argoproj.io"
	const apiVersion = "argoproj.io/v1alpha1"
	const kind = "Application"
	const project = "default"

	gitRepoURL := fmt.Sprintf("%v/%v/%v/%v.git",
		g.gitlabLib.GetSSHURL(ctx), g.clusterRepoConf.Parent.Path, params.Application.Name, params.Cluster)
	namespae := fmt.Sprintf("%v-%v", params.Environment, params.Application.GroupID)

	return &argocd.Application{
		APIVersion: apiVersion,
		Kind:       kind,
		Metadata: argocd.ApplicationMetadata{
			Finalizers: []string{finalizer},
			Name:       params.Cluster,
			Namespace:  argoCDNamespace,
		},
		Spec: argocd.ApplicationSpec{
			Source: argocd.ApplicationSource{
				RepoURL:        gitRepoURL,
				Path:           ".",
				TargetRevision: "master",
				Helm: &argocd.ApplicationSourceHelm{
					ValueFiles: []string{_filePathApplication, _filePathEnv, _filePathBase, _filePathSRE},
				},
			},
			Destination: argocd.ApplicationDestination{
				Server:    params.RegionEntity.K8SCluster.Server,
				Namespace: namespae,
			},
			Project: project,
			SyncPolicy: &argocd.SyncPolicy{
				SyncOptions: argocd.SyncOptions{"CreateNamespace=true"},
			},
		},
	}
}
