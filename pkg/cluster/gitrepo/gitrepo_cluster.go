package gitrepo

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"strings"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	"g.hz.netease.com/horizon/pkg/application/models"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	regionmodels "g.hz.netease.com/horizon/pkg/region/models"
	trmodels "g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"

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
                              │     ├── base.yaml             -- 基础数据
                              │     └── env.yaml              -- 环境相关数据
                              ├── pipeline
                              │     ├── pipeline.yaml         -- pipeline模板参数
                              │     └── pipeline-output.yaml  -- pipeline输出
							  └── argo
							        └── argo-application.yaml -- argo application
*/

const (
	_gitlabName = "compute"

	_branchMaster = "master"
	_branchGitops = "gitops"

	// _filePathChart           = "Chart.yaml"
	_filePathApplication = "application.yaml"
	_filePathSRE         = "sre/sre.yaml"
	_filePathBase        = "system/base.yaml"
	// _filePathEnv             = "system/env.yaml"
	_filePathPipeline = "pipeline/pipeline.yaml"
	// _filePathPipelineOutput  = "pipeline/pipeline-output.yaml"
	// _filePathArgoApplication = "argo/argo-application.yaml"
)

type Params struct {
	Cluster             string
	K8SServer           string
	HelmRepoURL         string
	Environment         string
	RegionEntity        regionmodels.RegionEntity
	PipelineJSONBlob    map[string]interface{}
	ApplicationJSONBlob map[string]interface{}
	TemplateRelease     *trmodels.TemplateRelease
	Application         *models.Application
}

type ClusterGitRepo interface {
	CreateCluster(ctx context.Context, params *Params) error
	UpdateCluster(ctx context.Context, params *Params) error
	DeleteCluster(ctx context.Context, cluster string, clusterID uint) error
	CompareConfig(ctx context.Context, cluster string) (string, error)
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

	// 1. create cluster repo
	if _, err := g.gitlabLib.CreateProject(ctx, params.Cluster, g.clusterRepoConf.Parent.ID); err != nil {
		return errors.E(op, err)
	}

	// 2. write files to repo
	pid := fmt.Sprintf("%v/%v", g.clusterRepoConf.Parent.Path, params.Cluster)
	// TODO(gjq) add Chart.yaml & argo-application.yaml
	var applicationYAML, pipelineYAML, baseValueYAML, sreValueYAML []byte
	var err1, err2, err3, err4 error
	marshal := func(b *[]byte, err *error, data interface{}) {
		*b, *err = yaml.Marshal(data)
	}
	marshal(&applicationYAML, &err1, g.assembleApplicationValue(params))
	marshal(&pipelineYAML, &err2, g.assemblePipelineValue(params))
	marshal(&baseValueYAML, &err3, g.assembleBaseValue(params))
	marshal(&sreValueYAML, &err4, g.assembleSREValue(params))

	for _, err := range []error{err1, err2, err3, err4} {
		if err != nil {
			return err
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
			FilePath: _filePathSRE,
			Content:  string(sreValueYAML),
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
	pid := fmt.Sprintf("%v/%v", g.clusterRepoConf.Parent.Path, params.Cluster)
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

func (g *clusterGitRepo) DeleteCluster(ctx context.Context, cluster string, clusterID uint) (err error) {
	const op = "cluster git repo: delete cluster"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. delete gitlab project
	pid := fmt.Sprintf("%v/%v", g.clusterRepoConf.Parent.Path, cluster)
	// 1.1 edit project's name and path to {cluster}-{clusterID}
	newName := fmt.Sprintf("%v-%d", cluster, clusterID)
	newPath := newName
	if err := g.gitlabLib.EditNameAndPathForProject(ctx, pid, &newName, &newPath); err != nil {
		return errors.E(op, err)
	}

	// 1.2 transfer project to RecyclingParent
	newPid := fmt.Sprintf("%v/%v", g.clusterRepoConf.Parent.Path, newPath)
	if err := g.gitlabLib.TransferProject(ctx, newPid, g.clusterRepoConf.RecyclingParent.Path); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *clusterGitRepo) CompareConfig(ctx context.Context, cluster string) (_ string, err error) {
	const op = "cluster git repo: compare config"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	pid := fmt.Sprintf("%v/%v", g.clusterRepoConf.Parent.Path, cluster)

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

type BaseValue struct {
	Application string `yaml:"application"`
	Cluster     string `yaml:"Cluster"`
	Environment string `yaml:"environment"`

	Region    string             `yaml:"region"`
	Group     uint               `yaml:"group"`
	Namespace string             `yaml:"namespace"`
	Template  *BaseValueTemplate `yaml:"template"`
	Priority  string             `yaml:"priority"`
	BaseImage *BaseValueImage    `yaml:"baseImage"`
}

type BaseValueTemplate struct {
	Name    string `yaml:"name"`
	Release string `yaml:"release"`
}

type BaseValueImage struct {
	Prefix string `yaml:"prefix"`
}

// assembleBaseValue assemble base value. return a map, key is template name, and value is *BaseValue
func (g *clusterGitRepo) assembleBaseValue(params *Params) map[string]*BaseValue {
	var namespace = fmt.Sprintf("%v-%v", params.Environment, params.Application.GroupID)
	ret := make(map[string]*BaseValue)
	ret[params.TemplateRelease.TemplateName] = &BaseValue{
		Application: params.Application.Name,
		Cluster:     params.Cluster,
		Environment: params.Environment,
		Region:      params.RegionEntity.Name,
		Group:       params.Application.GroupID,
		Namespace:   namespace,
		Template: &BaseValueTemplate{
			Name:    params.TemplateRelease.TemplateName,
			Release: params.TemplateRelease.Name,
		},
		Priority: string(params.Application.Priority),
		// TODO(gjq) cloudnative-template config
		BaseImage: &BaseValueImage{
			Prefix: path.Join(strings.TrimPrefix(strings.TrimPrefix(
				params.RegionEntity.Harbor.Server, "https://"), "http://"), "cloudnative-template"),
		},
	}
	return ret
}
