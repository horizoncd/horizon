package output

import (
	"errors"

	he "g.hz.netease.com/horizon/core/errors"
	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	perrors "g.hz.netease.com/horizon/pkg/errors"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"golang.org/x/net/context"
)

const _gitlabName = "control"

// Getter provides som functions for output
type Getter interface {
	// GetTemplateOutPut get horizon template's output for specified template release.
	GetTemplateOutPut(ctx context.Context, templateName, releaseName string) (string, error)
}

type getter struct {
	gitlabLib          gitlablib.Interface
	templateReleaseMgr manager.Manager
}

const (
	// output yaml file path
	_outputsPath = "output/outputs.yaml"
)

var (
	ErrGitlab = errors.New("GitlabError")
	ErrFormat = errors.New("FileFormatError")
)

func NewOutPutGetter(ctx context.Context, gitlabFty gitlabfty.Factory) (Getter, error) {
	gitlabLib, err := gitlabFty.GetByName(ctx, _gitlabName)
	if err != nil {
		return nil, err
	}
	return &getter{
		gitlabLib:          gitlabLib,
		templateReleaseMgr: manager.Mgr,
	}, nil
}

func (g *getter) GetTemplateOutPut(ctx context.Context,
	templateName, releaseName string) (string, error) {
	const op = "template output getter: getTemplateOutPut"
	defer wlog.Start(ctx, op).StopPrint()

	tr, err := g.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return "", err
	}

	bytes, err := g.gitlabLib.GetFile(ctx, tr.GitlabProject, tr.Name, _outputsPath)
	if err != nil {
		if _, ok := perrors.Cause(err).(*he.HorizonErrNotFound); !ok {
			log.Errorf(ctx, "Get Output file error, err = %s", err.Error())
			return "", err
		}
	}
	return string(bytes), nil
}
