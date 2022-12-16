package gitrepo

import (
	"context"
	"fmt"

	pkgcommon "github.com/horizoncd/horizon/pkg/common"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	gitlablib "github.com/horizoncd/horizon/lib/gitlab"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/util/angular"
	"github.com/horizoncd/horizon/pkg/util/log"
	"github.com/horizoncd/horizon/pkg/util/wlog"
	"github.com/xanzy/go-gitlab"
	"sigs.k8s.io/yaml"
)

const (
	_filePathManifest = "manifest.yaml"
	_branchMaster     = "master"

	_filePathApplication = "application.yaml"
	_filePathPipeline    = "pipeline.yaml"

	_applications          = "applications"
	_recyclingApplications = "recycling-applications"
)

type CreateOrUpdateRequest struct {
	Version string

	Environment  string
	BuildConf    map[string]interface{}
	TemplateConf map[string]interface{}
}

type GetResponse struct {
	Manifest     map[string]interface{}
	BuildConf    map[string]interface{}
	TemplateConf map[string]interface{}
}

type ApplicationGitRepo interface {
	CreateOrUpdateApplication(ctx context.Context, application string, request CreateOrUpdateRequest) error
	GetApplication(ctx context.Context, application, environment string) (*GetResponse, error)
	// HardDeleteApplication hard delete an application by the specified application name
	HardDeleteApplication(ctx context.Context, application string) error
}

type gitRepo struct {
	gitlabLib                  gitlablib.Interface
	applicationsGroup          *gitlab.Group
	recyclingApplicationsGroup *gitlab.Group
}

var _ ApplicationGitRepo = &gitRepo{}

func NewApplicationGitlabRepo(ctx context.Context, rootGroup *gitlab.Group,
	gitlabLib gitlablib.Interface) (ApplicationGitRepo, error) {
	applicationsGroup, err := gitlabLib.GetCreatedGroup(ctx, rootGroup.ID, rootGroup.FullPath, _applications)
	if err != nil {
		return nil, err
	}
	recyclingApplicationsGroup, err := gitlabLib.GetCreatedGroup(ctx, rootGroup.ID,
		rootGroup.FullPath, _recyclingApplications)
	if err != nil {
		return nil, err
	}
	return &gitRepo{
		gitlabLib:                  gitlabLib,
		applicationsGroup:          applicationsGroup,
		recyclingApplicationsGroup: recyclingApplicationsGroup,
	}, nil
}

func (g gitRepo) CreateOrUpdateApplication(ctx context.Context, application string, req CreateOrUpdateRequest) error {
	const op = "gitlab repo: create or update application"
	defer wlog.Start(ctx, op).StopPrint()

	currentUser, err := common.UserFromContext(ctx)
	if err != nil {
		return err
	}

	environmentRepoName := common.ApplicationRepoDefaultEnv
	if req.Environment != "" {
		environmentRepoName = req.Environment
	}

	var envProjectExists = false
	pid := fmt.Sprintf("%v/%v/%v", g.applicationsGroup.FullPath, application, environmentRepoName)
	_, err = g.gitlabLib.GetProject(ctx, pid)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
			return err
		}
		// if not found, test application group exist
		gid := fmt.Sprintf("%v/%v", g.applicationsGroup.FullPath, application)
		parentGroup, err := g.gitlabLib.GetGroup(ctx, gid)
		if err != nil {
			if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); !ok {
				return err
			}
			parentGroup, err = g.gitlabLib.CreateGroup(ctx, application, application, &g.applicationsGroup.ID)
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
		manifest := pkgcommon.Manifest{Version: req.Version}
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

func (g gitRepo) GetApplication(ctx context.Context, application, environment string) (*GetResponse, error) {
	const op = "gitlab repo: get application"
	defer wlog.Start(ctx, op).StopPrint()

	// 1. get data from gitlab
	gid := fmt.Sprintf("%v/%v", g.applicationsGroup.FullPath, application)
	pid := fmt.Sprintf("%v/%v", gid, func() string {
		if environment == "" {
			return common.ApplicationRepoDefaultEnv
		}
		return environment
	}())

	// if env template not exist, use the default one
	_, err := g.gitlabLib.GetProject(ctx, pid)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			pid = fmt.Sprintf("%v/%v", gid, common.ApplicationRepoDefaultEnv)
		}
	}

	manifestBytes, err1 := g.gitlabLib.GetFile(ctx, pid, _branchMaster, _filePathManifest)
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
		var entity map[string]interface{}
		err = yaml.Unmarshal(bytes, &entity)
		if err != nil {
			return nil, perror.Wrap(herrors.ErrParamInvalid, err.Error())
		}
		return entity, nil
	}

	if manifestBytes != nil {
		entity, err := TransformData(manifestBytes)
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

func (g gitRepo) HardDeleteApplication(ctx context.Context, application string) error {
	const op = "gitlab repo: hard delete application"
	defer wlog.Start(ctx, op).StopPrint()

	gid := fmt.Sprintf("%v/%v", g.applicationsGroup.FullPath, application)
	return g.gitlabLib.DeleteGroup(ctx, gid)
}
