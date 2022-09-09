package gitrepo

import (
	"context"
	"encoding/json"
	"fmt"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"sigs.k8s.io/yaml"
	kyaml "sigs.k8s.io/yaml"
)

const (
	_filePathManifest = "manifest.yaml"
)

type Manifest struct {
	Version string `yaml:"version"`
}

type CreateOrUpdateRequest struct {
	Version string
	// TODO(encode the template info into manifest)
	Environment  string
	BuildConf    map[string]interface{}
	TemplateConf map[string]interface{}
}

type GetResponse struct {
	Manifest     map[string]interface{}
	BuildConf    map[string]interface{}
	TemplateConf map[string]interface{}
}

const (
	_branchMaster = "master"

	_filePathApplication = "application.yaml"
	_filePathPipeline    = "pipeline.yaml"

	_default = "default"
)

type ApplicationGitRepo2 interface {
	CreateOrUpdateApplication(ctx context.Context, application string, request CreateOrUpdateRequest) error
	GetApplication(ctx context.Context, application, environment string) (*GetResponse, error)
	// DeleteApplication soft delete an application by the specified application name
	DeleteApplication(ctx context.Context, application string, applicationID uint) error
	// HardDeleteApplication hard delete an application by the specified application name
	HardDeleteApplication(ctx context.Context, application string) error
}

type gitRepo2 struct {
	gitlabLib gitlablib.Interface
	repoConf  *gitlabconf.Repo
}

var _ ApplicationGitRepo2 = &gitRepo2{}

func NewApplicationGitlabRepo2(ctx context.Context, gitlabRepoConfig gitlabconf.RepoConfig,
	gitlabLib gitlablib.Interface) (ApplicationGitRepo2, error) {
	return &gitRepo2{
		gitlabLib: gitlabLib,
		repoConf:  gitlabRepoConfig.Application,
	}, nil
}

func (g gitRepo2) CreateOrUpdateApplication(ctx context.Context, application string, req CreateOrUpdateRequest) error {
	const op = "gitlab repo: create or update application"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	environmentRepoName := _default
	if req.Environment != "" {
		environmentRepoName = req.Environment
	}

	var envProjectExists = false
	pid := fmt.Sprintf("%v/%v/%v", g.repoConf.Parent.Path, application, environmentRepoName)
	_, err = g.gitlabLib.GetProject(ctx, pid)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
		// if not found, test application group exist
		gid := fmt.Sprintf("%v/%v", g.repoConf.Parent.Path, application)
		parentGroup, err := g.gitlabLib.GetGroup(ctx, gid)
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
				return err
			}
			parentGroup, err = g.gitlabLib.CreateGroup(ctx, application, application, &g.repoConf.Parent.ID)
			if err != nil {
				return err
			}
		}
		_, err = g.gitlabLib.CreateProject(ctx, environmentRepoName, parentGroup.ID)
		if err != nil {
			return err
		}
	} else {
		envProjectExists = true
	}

	// 2. if env template repo exists, the gitlab action is update, else the action is create
	var action = gitlablib.FileCreate
	if envProjectExists {
		action = gitlablib.FileUpdate
	}

	// 3. write files
	var templateConfYaml, buildConfYaml, manifestYaml []byte
	if req.TemplateConf != nil {
		templateConfYaml, err = yaml.Marshal(req.TemplateConf)
		if err != nil {
			log.Warningf(ctx, "templateConf marshal error, %v", req.TemplateConf)
			return perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
	}
	if req.BuildConf != nil {
		buildConfYaml, err = yaml.Marshal(req.BuildConf)
		if err != nil {
			log.Warningf(ctx, "buildConf marshal error, %v", req.BuildConf)
			return perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
	}
	if req.Version != "" {
		manifest := Manifest{Version: req.Version}
		manifestYaml, err = yaml.Marshal(manifest)
		if err != nil {
			log.Warningf(ctx, "Manifest marshal error, %+v", manifest)
			return perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
	}

	actions := func() []gitlablib.CommitAction {
		actions := make([]gitlablib.CommitAction, 0)
		if req.BuildConf != nil {
			actions = append(actions, gitlablib.CommitAction{
				Action:   action,
				FilePath: _filePathPipeline,
				Content:  string(buildConfYaml),
			})
		}
		if req.TemplateConf != nil {
			actions = append(actions, gitlablib.CommitAction{
				Action:   action,
				FilePath: _filePathApplication,
				Content:  string(templateConfYaml),
			})
		}
		if req.Version != "" {
			actions = append(actions, gitlablib.CommitAction{
				Action:   action,
				FilePath: _filePathManifest,
				Content:  string(manifestYaml),
			})
		}
		return actions
	}()

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.GetName(),
		Action:      fmt.Sprintf("%s application %s configure", string(action), environmentRepoName),
		Application: angular.StringPtr(application),
	}, struct {
		Application map[string]interface{} `json:"application"`
		Pipeline    map[string]interface{} `json:"pipeline"`
	}{
		Application: req.TemplateConf,
		Pipeline:    req.BuildConf,
	})
	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return err
	}
	return nil
}

func (g gitRepo2) GetApplication(ctx context.Context, application, environment string) (*GetResponse, error) {
	const op = "gitlab repo: get application"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get data from gitlab
	gid := fmt.Sprintf("%v/%v", g.repoConf.Parent.Path, application)
	pid := fmt.Sprintf("%v/%v", gid, func() string {
		if environment == "" {
			return _default
		}
		return environment
	}())
	manifestbytes, err1 := g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathManifest)
	buildConfBytes, err2 := g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathPipeline)
	templateConfBytes, err3 := g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathApplication)
	for _, err := range []error{err1, err2, err3} {
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
				return nil, err
			}
		}
	}

	// 2. process data
	res := GetResponse{}
	TransformData := func(bytes []byte) (map[string]interface{}, error) {
		jsonBytes, err := kyaml.YAMLToJSON(bytes)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
		var entity map[string]interface{}
		err = json.Unmarshal(jsonBytes, &entity)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
		return entity, nil
	}

	if manifestbytes != nil {
		entity, err := TransformData(manifestbytes)
		if err != nil {
			return nil, err
		}
		res.Manifest = entity
	}

	if buildConfBytes != nil {
		entity, err := TransformData(buildConfBytes)
		if err != nil {
			return nil, err
		}
		res.BuildConf = entity
	}

	if templateConfBytes != nil {
		entity, err := TransformData(templateConfBytes)
		if err != nil {
			return nil, err
		}
		res.TemplateConf = entity
	}
	return &res, nil
}

func (g gitRepo2) DeleteApplication(ctx context.Context, application string, applicationID uint) error {
	const op = "gitlab repo: hard delete application"
	defer wlog.Start(ctx, op).StopPrint()

	gid := fmt.Sprintf("%v/%v", g.repoConf.Parent.Path, application)
	return g.gitlabLib.DeleteGroup(ctx, gid)
}

func (g gitRepo2) HardDeleteApplication(ctx context.Context, application string) error {
	const op = "gitlab repo: hard delete application"
	defer wlog.Start(ctx, op).StopPrint()

	gid := fmt.Sprintf("%v/%v", g.repoConf.Parent.Path, application)
	return g.gitlabLib.DeleteGroup(ctx, gid)
}
