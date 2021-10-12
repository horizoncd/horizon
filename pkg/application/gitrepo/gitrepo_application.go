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
	gitlabsvc "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	_branchMaster = "master"

	_filePathApplication = "application.json"
	_filePathPipeline    = "pipeline.json"
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
	gitlabFactory gitlabsvc.Factory
	gitlabConfig  gitlabconf.Config
}

func NewApplicationGitlabRepo(gitlabConfig gitlabconf.Config) ApplicationGitRepo {
	return &applicationGitlabRepo{
		gitlabFactory: gitlabsvc.Fty,
		gitlabConfig:  gitlabConfig,
	}
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

	var applicationConf = g.gitlabConfig.Application

	gitlabLib, err := g.gitlabFactory.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 1. create application repo
	if _, err := gitlabLib.CreateProject(ctx, application, applicationConf.Parent.ID); err != nil {
		return errors.E(op, err)
	}

	// 2. write files to application repo
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, application)
	applicationJSON, err := json.MarshalIndent(applicationJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	pipelineJSON, err := json.MarshalIndent(pipelineJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileCreate,
			FilePath: _filePathApplication,
			Content:  string(applicationJSON),
		}, {
			Action:   gitlablib.FileCreate,
			FilePath: _filePathPipeline,
			Content:  string(pipelineJSON),
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

	if _, err := gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
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

	var applicationConf = g.gitlabConfig.Application

	gitlabLib, err := g.gitlabFactory.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 3. write files to gitlab
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, application)
	applicationJSON, err := json.MarshalIndent(applicationJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	pipelineJSON, err := json.MarshalIndent(pipelineJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathApplication,
			Content:  string(applicationJSON),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathPipeline,
			Content:  string(pipelineJSON),
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

	if _, err := gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *applicationGitlabRepo) GetApplication(ctx context.Context,
	application string) (pipelineJSONBlob, applicationJSONBlob map[string]interface{}, err error) {
	const op = "gitlab repo: get application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	var applicationConf = g.gitlabConfig.Application

	gitlabLib, err := g.gitlabFactory.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return nil, nil, errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 2. get template and pipeline from gitlab
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, application)
	var applicationBytes, pipelineBytes []byte
	var err1, err2 error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		pipelineBytes, err1 = gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathPipeline)
	}()
	go func() {
		defer wg.Done()
		applicationBytes, err2 = gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathApplication)
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
	if err := json.Unmarshal(applicationBytes, &pipelineJSONBlob); err != nil {
		return nil, nil, errors.E(op, err)
	}

	return pipelineJSONBlob, applicationJSONBlob, nil
}

func (g *applicationGitlabRepo) DeleteApplication(ctx context.Context,
	application string, applicationID uint) (err error) {
	const op = "gitlab repo: get application"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	var applicationConf = g.gitlabConfig.Application

	// 1. get gitlab lib instance
	gitlabLib, err := g.gitlabFactory.GetByName(ctx, applicationConf.GitlabName)
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}

	// 2. delete gitlab project
	pid := fmt.Sprintf("%v/%v", applicationConf.Parent.Path, application)
	// 2.1 transfer project to RecyclingParent
	if err := gitlabLib.TransferProject(ctx, pid, applicationConf.RecyclingParent.Path); err != nil {
		return errors.E(op, err)
	}
	// 2.2 edit project's name and path to {application}-{applicationID}
	newPid := fmt.Sprintf("%v/%v", applicationConf.RecyclingParent.Path, application)
	newName := fmt.Sprintf("%v-%d", application, applicationID)
	newPath := newName
	if err := gitlabLib.EditNameAndPathForProject(ctx, newPid, &newName, &newPath); err != nil {
		return errors.E(op, err)
	}

	return nil
}
