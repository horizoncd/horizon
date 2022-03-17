package gitrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/core/middleware/user"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	perror "g.hz.netease.com/horizon/pkg/errors"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/wlog"

	"github.com/xanzy/go-gitlab"
	"gopkg.in/yaml.v2"
	kyaml "sigs.k8s.io/yaml"
)

const (
	_gitlabName = "compute"

	_branchMaster = "master"

	_filePathApplication = "application.yaml"
	_filePathPipeline    = "pipeline.yaml"

	_default = "default"
)

// ApplicationGitRepo interface to provide the management functions with git repo for applications
type ApplicationGitRepo interface {
	// CreateApplication create an application, with the pipeline jsonBlob and application jsonBlob
	CreateApplication(ctx context.Context, application string,
		pipelineJSONBlob, applicationJSONBlob map[string]interface{}) error
	// UpdateApplication update an application, with the updated pipeline jsonBlob and updated application jsonBlob
	UpdateApplication(ctx context.Context, application string,
		pipelineJSONBlob, applicationJSONBlob map[string]interface{}) error
	// GetApplication get an application, return the pipeline jsonBlob and application jsonBlob
	GetApplication(ctx context.Context, application string) (pipelineJSONBlob,
		applicationJSONBlob map[string]interface{}, err error)
	// UpdateApplicationEnvTemplate update application's env template
	UpdateApplicationEnvTemplate(ctx context.Context, application, env string,
		pipelineJSONBlob, applicationJSONBlob map[string]interface{}) error
	// GetApplicationEnvTemplate get application's env template
	// if env template is not exists, return the default template
	GetApplicationEnvTemplate(ctx context.Context, application, env string) (pipelineJSONBlob,
		applicationJSONBlob map[string]interface{}, err error)
	// DeleteApplication delete an application by the specified application name
	DeleteApplication(ctx context.Context, application string, applicationID uint) error
}

type applicationGitlabRepo struct {
	gitlabLib           gitlablib.Interface
	applicationRepoConf *gitlabconf.Repo
}

func NewApplicationGitlabRepo(ctx context.Context, gitlabRepoConfig gitlabconf.RepoConfig,
	gitlabFactory gitlabfty.Factory) (ApplicationGitRepo, error) {
	gitlabLib, err := gitlabFactory.GetByName(ctx, _gitlabName)
	if err != nil {
		return nil, err
	}
	return &applicationGitlabRepo{
		gitlabLib:           gitlabLib,
		applicationRepoConf: gitlabRepoConfig.Application,
	}, nil
}

func (g *applicationGitlabRepo) CreateApplication(ctx context.Context, application string,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}) (err error) {
	const op = "gitlab repo: create application"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. create application group
	group, err := g.gitlabLib.CreateGroup(ctx, application, application, &g.applicationRepoConf.Parent.ID)
	if err != nil {
		return err
	}

	// 2. create application default repo
	if _, err := g.gitlabLib.CreateProject(ctx, _default, group.ID); err != nil {
		return err
	}

	// 3. write files
	return g.createOrUpdateApplication(ctx, application, _default, gitlablib.FileCreate,
		pipelineJSONBlob, applicationJSONBlob)
}

func (g *applicationGitlabRepo) UpdateApplication(ctx context.Context, application string,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}) (err error) {
	return g.createOrUpdateApplication(ctx, application, _default, gitlablib.FileUpdate,
		pipelineJSONBlob, applicationJSONBlob)
}

func (g *applicationGitlabRepo) GetApplication(ctx context.Context,
	application string) (pipelineJSONBlob, applicationJSONBlob map[string]interface{}, err error) {
	return g.getApplication(ctx, application, _default)
}

func (g *applicationGitlabRepo) UpdateApplicationEnvTemplate(ctx context.Context, application, env string,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}) (err error) {
	const op = "gitlab repo: update application env template"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. check env template repo exists
	var envProjectExists = false
	pid := fmt.Sprintf("%v/%v/%v", g.applicationRepoConf.Parent.Path, application, env)
	_, err = g.gitlabLib.GetProject(ctx, pid)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
		// if not found, create this repo first
		gid := fmt.Sprintf("%v/%v", g.applicationRepoConf.Parent.Path, application)
		parentGroup, err := g.gitlabLib.GetGroup(ctx, gid)
		if err != nil {
			return err
		}
		_, err = g.gitlabLib.CreateProject(ctx, env, parentGroup.ID)
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
	return g.createOrUpdateApplication(ctx, application, env, action, pipelineJSONBlob, applicationJSONBlob)
}

func (g *applicationGitlabRepo) GetApplicationEnvTemplate(ctx context.Context,
	application, env string) (pipelineJSONBlob, applicationJSONBlob map[string]interface{}, err error) {
	// 1. check env template repo exists
	pid := fmt.Sprintf("%v/%v/%v", g.applicationRepoConf.Parent.Path, application, env)
	_, err = g.gitlabLib.GetProject(ctx, pid)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return nil, nil, err
		}
		// if not found, return the default template
		return g.getApplication(ctx, application, _default)
	}

	// 2. if env template repo exists, return the env template
	return g.getApplication(ctx, application, env)
}

func (g *applicationGitlabRepo) DeleteApplication(ctx context.Context,
	application string, applicationID uint) (err error) {
	const op = "gitlab repo: delete application"
	defer wlog.Start(ctx, op).StopPrint()

	// page and perPage used for list projects.
	// the count of project under the application's group cannot be greater than 8,
	// because we only have 7 envs: dev, test, reg, perf, pre, beta, online
	// so the default page is 1, and the default perPage is 8
	const (
		page    = 1
		perPage = 8
	)

	gid := fmt.Sprintf("%v/%v", g.applicationRepoConf.Parent.Path, application)

	recyclingGroupName := fmt.Sprintf("%v-%d", application, applicationID)

	// 1. create recyclingGroup
	recyclingGroup, err := g.gitlabLib.CreateGroup(ctx, recyclingGroupName,
		recyclingGroupName, &g.applicationRepoConf.RecyclingParent.ID)
	if err != nil {
		return err
	}

	// 2. transfer all project to recyclingGroup
	projects, err := g.gitlabLib.ListGroupProjects(ctx, gid, page, perPage)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	var errs []error
	for _, project := range projects {
		wg.Add(1)
		go func(project *gitlab.Project) {
			defer wg.Done()
			if err := g.gitlabLib.TransferProject(ctx, project.ID, recyclingGroup.ID); err != nil {
				errs = append(errs, err)
			}
		}(project)
	}
	wg.Wait()

	if len(errs) > 0 {
		return err
	}

	// 3. delete old application group
	return g.gitlabLib.DeleteGroup(ctx, gid)
}

func (g *applicationGitlabRepo) createOrUpdateApplication(ctx context.Context, application, repo string,
	action gitlablib.FileAction, pipelineJSONBlob, applicationJSONBlob map[string]interface{}) error {
	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return err
	}

	pid := fmt.Sprintf("%v/%v/%v", g.applicationRepoConf.Parent.Path, application, repo)

	// 2. write files to gitlab
	applicationYAML, err := yaml.Marshal(applicationJSONBlob)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	pipelineYAML, err := yaml.Marshal(pipelineJSONBlob)
	if err != nil {
		return perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	actions := []gitlablib.CommitAction{
		{
			Action:   action,
			FilePath: _filePathApplication,
			Content:  string(applicationYAML),
		}, {
			Action:   action,
			FilePath: _filePathPipeline,
			Content:  string(pipelineYAML),
		},
	}

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.GetName(),
		Action:      fmt.Sprintf("%s application %s template", string(action), repo),
		Application: angular.StringPtr(application),
	}, struct {
		Application map[string]interface{} `json:"application"`
		Pipeline    map[string]interface{} `json:"pipeline"`
	}{
		Application: applicationJSONBlob,
		Pipeline:    pipelineJSONBlob,
	})

	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return err
	}

	return nil
}

func (g *applicationGitlabRepo) getApplication(ctx context.Context,
	application, repo string) (pipelineJSONBlob, applicationJSONBlob map[string]interface{}, err error) {
	const op = "gitlab repo: get application"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get template and pipeline from gitlab
	gid := fmt.Sprintf("%v/%v", g.applicationRepoConf.Parent.Path, application)
	pid := fmt.Sprintf("%v/%v", gid, repo)

	var applicationBytes, pipelineBytes []byte
	var err1, err2 error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pipelineBytes, err1 = g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathPipeline)
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
		applicationBytes, err2 = g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathApplication)
		if err2 != nil {
			return
		}
		applicationBytes, err2 = kyaml.YAMLToJSON(applicationBytes)
		if err2 != nil {
			err2 = perror.Wrap(herrors.ErrParamInvalid, err2.Error())
		}
	}()
	wg.Wait()

	for _, err := range []error{err1, err2} {
		if err != nil {
			return nil, nil, err
		}
	}

	if err := json.Unmarshal(pipelineBytes, &pipelineJSONBlob); err != nil {
		return nil, nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}
	if err := json.Unmarshal(applicationBytes, &applicationJSONBlob); err != nil {
		return nil, nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
	}

	return pipelineJSONBlob, applicationJSONBlob, nil
}
