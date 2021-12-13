package output

import (
	"errors"

	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	gitlabfty "g.hz.netease.com/horizon/pkg/gitlab/factory"
	"g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/util/log"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"golang.org/x/net/context"
	"gopkg.in/yaml.v2"
)

const _gitlabName = "control"

// Getter provides som functions for output
type Getter interface {
	// GetTemplateOutPut get horizon template's output for specified template release.
	GetTemplateOutPut(ctx context.Context, templateName, releaseName string) (map[string]interface{}, error)
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
		gitlabLib: gitlabLib,
	}, nil
}

func (g *getter) GetTemplateOutPut(ctx context.Context,
	templateName, releaseName string) (map[string]interface{}, error) {
	const op = "template output getter: getTemplateOutPut"
	defer wlog.Start(ctx, op).StopPrint()

	tr, err := g.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}

	bytes, err := g.gitlabLib.GetFile(ctx, tr.GitlabProject, tr.Name, _outputsPath)
	if err != nil {
		log.Errorf(ctx, "Get Output file error, err = %s", err.Error())
		return nil, ErrGitlab
	}

	outPutMap := make(map[interface{}]interface{})
	err = yaml.Unmarshal(bytes, &outPutMap)
	if err != nil {
		log.Errorf(ctx, "yaml unmarshal error, err = %s", err.Error())
		return nil, ErrFormat
	}

	return nil, nil
}
