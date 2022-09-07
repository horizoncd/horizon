package gitrepo

import (
	"context"
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
)

type CreateOrUpdateRequest struct {
	Version      string
	Environment  *string
	BuildConf    map[string]interface{}
	TemplateConf map[string]interface{}
}

type GetResponse struct {
	Version                 string
	PipelineJson            map[string]interface{}
	TemplateConfigJsonInput map[string]interface{}
}

type GitRepoInterface2 interface {
	CreateOrUpdateApplication(ctx context.Context, application string, request CreateOrUpdateRequest) error
	GetApplication(ctx context.Context, application, environment string) (GetResponse, error)
	// DeleteApplication soft delete an application by the specified application name
	DeleteApplication(ctx context.Context, application string, applicationID uint) error
	// HardDeleteApplication hard delete an application by the specified application name
	HardDeleteApplication(ctx context.Context, application string) error
}

type gitRepo2 struct {
	gitlabLib gitlablib.Interface
	repoConf  *gitlabconf.Repo
}

var _ GitRepoInterface2 = &gitRepo2{}

func NewGitRepo2(ctx context.Context, gitlabRepoConfig gitlabconf.RepoConfig,
	gitlabLib gitlablib.Interface) (GitRepoInterface2, error) {
	return &gitRepo2{
		gitlabLib: gitlabLib,
		repoConf:  gitlabRepoConfig.Application,
	}, nil
}

func (g gitRepo2) CreateOrUpdateApplication(ctx context.Context, application string, req CreateOrUpdateRequest) error {
	const op = "gitlab repo: create application"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	environmentRepoName := _default
	if req.Environment != nil {
		environmentRepoName = *req.Environment
	}

	var envProjectExists = false
	pid := fmt.Sprintf("%v/%v/%v", g.repoConf.Parent.Path, application, environmentRepoName)
	_, err = g.gitlabLib.GetProject(ctx, pid)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
		// if not found, create this repo first
		gid := fmt.Sprintf("%v/%v", g.repoConf.Parent.Path, application)
		parentGroup, err := g.gitlabLib.GetGroup(ctx, gid)
		if err != nil {
			return err
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
	templateConfYAML, err := yaml.Marshal(req.TemplateConf)
	if err != nil {
		log.Warningf(ctx, "templateConf marshal error, %v", req.TemplateConf)
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	buildConfYaml, err := yaml.Marshal(req.BuildConf)
	if err != nil {
		log.Warningf(ctx, "buildConf marshal error, %v", req.BuildConf)
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
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
				Content:  string(templateConfYAML),
			})
		}
		// TODO(tom) add version file here
		return actions
	}()

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.GetName(),
		Action:      fmt.Sprintf("%s application %s template", string(action), environmentRepoName),
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

func (g gitRepo2) createOrUpdateApplication(ctx context.Context, application,
	repo string, request CreateOrUpdateRequest) error {

}

func (g gitRepo2) GetApplication(ctx context.Context, application, environment string) (GetResponse, error) {
	panic("implement me")
}

func (g gitRepo2) DeleteApplication(ctx context.Context, application string, applicationID uint) error {
	panic("implement me")
}

func (g gitRepo2) HardDeleteApplication(ctx context.Context, application string) error {
	panic("implement me")
}
