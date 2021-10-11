package gitrepo

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/middleware/user"
	gitlabconf "g.hz.netease.com/horizon/pkg/config/gitlab"
	gitlablib "g.hz.netease.com/horizon/pkg/lib/gitlab"
	gitlabsvc "g.hz.netease.com/horizon/pkg/service/gitlab"
	"g.hz.netease.com/horizon/pkg/util/angular"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

const (
	_branchMaster = "master"

	_filePathCD = "cd.json"
	_filePathCI = "ci.json"
)

// ApplicationGitRepo interface to provide the management functions with git repo for applications
type ApplicationGitRepo interface {
	// CreateApplication create an application, with the ci formData and cd formData
	CreateApplication(ctx context.Context, application string, ciJSONBlob, cdJSONBlob map[string]interface{}) error
	// UpdateApplication update an application, with the updated ci formData and updated cd formData
	UpdateApplication(ctx context.Context, application string, ciJSONBlob, cdJSONBlob map[string]interface{}) error
	// GetApplication get an application, return the ci formData and cd formData
	GetApplication(ctx context.Context, application string) (ciJSONBlob, cdJSONBlob map[string]interface{}, err error)
	// DeleteApplication delete an application by the specified application name
	DeleteApplication(ctx context.Context, application string) error
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
	ciJSONBlob, cdJSONBlob map[string]interface{}) (err error) {
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
	cdJSON, err := json.MarshalIndent(cdJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	ciJSON, err := json.MarshalIndent(ciJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileCreate,
			FilePath: _filePathCD,
			Content:  string(cdJSON),
		}, {
			Action:   gitlablib.FileCreate,
			FilePath: _filePathCI,
			Content:  string(ciJSON),
		},
	}

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.GetName(),
		Action:      "create application",
		Application: angular.StringPtr(application),
	}, struct {
		CD map[string]interface{} `json:"cd"`
		CI map[string]interface{} `json:"ci"`
	}{
		CD: cdJSONBlob,
		CI: ciJSONBlob,
	})

	if _, err := gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *applicationGitlabRepo) UpdateApplication(ctx context.Context, application string,
	ciJSONBlob, cdJSONBlob map[string]interface{}) (err error) {
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
	cdJSON, err := json.MarshalIndent(cdJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	ciJSON, err := json.MarshalIndent(ciJSONBlob, "", "  ")
	if err != nil {
		return errors.E(op, http.StatusInternalServerError,
			errors.ErrorCode(common.InternalError), err)
	}
	actions := []gitlablib.CommitAction{
		{
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathCD,
			Content:  string(cdJSON),
		}, {
			Action:   gitlablib.FileUpdate,
			FilePath: _filePathCI,
			Content:  string(ciJSON),
		},
	}

	commitMsg := angular.CommitMessage("application", angular.Subject{
		Operator:    currentUser.GetName(),
		Action:      "update application",
		Application: angular.StringPtr(application),
	}, struct {
		CD map[string]interface{} `json:"cd"`
		CI map[string]interface{} `json:"ci"`
	}{
		CD: cdJSONBlob,
		CI: ciJSONBlob,
	})

	if _, err := gitlabLib.WriteFiles(ctx, pid, _branchMaster, commitMsg, nil, actions); err != nil {
		return errors.E(op, err)
	}

	return nil
}

func (g *applicationGitlabRepo) GetApplication(ctx context.Context,
	application string) (ciJSONBlob, cdJSONBlob map[string]interface{}, err error) {
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
	var cdBytes, ciBytes []byte
	var err1, err2 error

	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		defer wg.Done()
		ciBytes, err1 = gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathCI)
	}()
	go func() {
		defer wg.Done()
		cdBytes, err2 = gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathCD)
	}()
	wg.Wait()

	for _, err := range []error{err1, err2} {
		if err != nil {
			return nil, nil, errors.E(op, err)
		}
	}

	if err := json.Unmarshal(ciBytes, &ciJSONBlob); err != nil {
		return nil, nil, errors.E(op, err)
	}
	if err := json.Unmarshal(cdBytes, &ciJSONBlob); err != nil {
		return nil, nil, errors.E(op, err)
	}

	return ciJSONBlob, cdJSONBlob, nil
}

func (g *applicationGitlabRepo) DeleteApplication(ctx context.Context, application string) (err error) {
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
	if err := gitlabLib.DeleteProject(ctx, pid); err != nil {
		return errors.E(op, err)
	}

	return nil
}
