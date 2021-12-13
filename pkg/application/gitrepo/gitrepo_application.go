package gitrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/errors"
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
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. create application group
	group, err := g.gitlabLib.CreateGroup(ctx, application, application, &g.applicationRepoConf.Parent.ID)
	if err != nil {
		return errors.E(op, err)
	}

	// 2. create application default repo
	if _, err := g.gitlabLib.CreateProject(ctx, _default, group.ID); err != nil {
		return errors.E(op, err)
	}

	// 3. write files to application repo
	pid := fmt.Sprintf("%v/%v/%v", g.applicationRepoConf.Parent.Path, application, _default)
	applicationYAML, err := yaml.Marshal(applicationJSONBlob)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	pipelineYAML, err := yaml.Marshal(pipelineJSONBlob)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
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
		},
	}

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.GetName(),
		Action:      "create application",
		Application: angular.StringPtr(application),
	}, struct {
		Application map[string]interface{} `json:"application"`
		Pipeline    map[string]interface{} `json:"pipeline"`
	}{
		Application: applicationJSONBlob,
		Pipeline:    pipelineJSONBlob,
	})

	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *applicationGitlabRepo) UpdateApplication(ctx context.Context, application string,
	pipelineJSONBlob, applicationJSONBlob map[string]interface{}) (err error) {
	const op = "gitlab repo: update application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	currentUser, err := user.FromContext(ctx)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), "no user in context")
	}

	// 1. write files to gitlab
	pid := fmt.Sprintf("%v/%v/%v", g.applicationRepoConf.Parent.Path, application, _default)
	applicationYAML, err := yaml.Marshal(applicationJSONBlob)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	pipelineYAML, err := yaml.Marshal(pipelineJSONBlob)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
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
		},
	}

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.GetName(),
		Action:      "update application",
		Application: angular.StringPtr(application),
	}, struct {
		Application map[string]interface{} `json:"application"`
		Pipeline    map[string]interface{} `json:"pipeline"`
	}{
		Application: applicationJSONBlob,
		Pipeline:    pipelineJSONBlob,
	})

	if _, err := g.gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *applicationGitlabRepo) GetApplication(ctx context.Context,
	application string) (pipelineJSONBlob, applicationJSONBlob map[string]interface{}, err error) {
	const op = "gitlab repo: get application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	// 1. get template and pipeline from gitlab
	gid := fmt.Sprintf("%v/%v", g.applicationRepoConf.Parent.Path, application)
	pid := fmt.Sprintf("%v/%v", gid, _default)

	// TODO(gjq): after migration, remove this check
	// check if '{application}/default' project exists
	_, err = g.gitlabLib.GetProject(ctx, pid)
	if err != nil {
		if errors.Status(err) != http.StatusNotFound {
			return nil, nil, errors.E(op, err)
		}
		// if not found, use '{application}', pid = gid
		pid = gid
	}

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
	}()
	go func() {
		defer wg.Done()
		applicationBytes, err2 = g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathApplication)
		if err2 != nil {
			return
		}
		applicationBytes, err2 = kyaml.YAMLToJSON(applicationBytes)
	}()
	wg.Wait()

	for _, err := range []error{err1, err2} {
		if err != nil {
			return nil, nil, errors.E(op, err)
		}
	}

	if err := json.Unmarshal(pipelineBytes, &pipelineJSONBlob); err != nil {
		return nil, nil, errors.E(op, err)
	}
	if err := json.Unmarshal(applicationBytes, &applicationJSONBlob); err != nil {
		return nil, nil, errors.E(op, err)
	}

	return pipelineJSONBlob, applicationJSONBlob, nil
}

func (g *applicationGitlabRepo) DeleteApplication(ctx context.Context,
	application string, applicationID uint) (err error) {
	const op = "gitlab repo: delete application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

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
		return errors.E(op, err)
	}

	// 2. transfer all project to recyclingGroup
	projects, err := g.gitlabLib.ListGroupProjects(ctx, gid, page, perPage)
	if err != nil {
		return errors.E(op, err)
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
		return errors.E(op, errs[0])
	}

	// 3. delete old application group
	if err := g.gitlabLib.DeleteGroup(ctx, gid); err != nil {
		return errors.E(op, err)
	}

	return nil
}
