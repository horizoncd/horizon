package output

import (
	"errors"

	"g.hz.netease.com/horizon/pkg/param/managerparam"
	tmanager "g.hz.netease.com/horizon/pkg/template/manager"
	"g.hz.netease.com/horizon/pkg/templaterelease/manager"
	"g.hz.netease.com/horizon/pkg/templaterepo"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"golang.org/x/net/context"
)

// Getter provides som functions for output
type Getter interface {
	// GetTemplateOutPut get horizon template's output for specified template release.
	GetTemplateOutPut(ctx context.Context, templateName, releaseName string) (string, error)
}

type getter struct {
	templateRepo       templaterepo.TemplateRepo
	templateMgr        tmanager.Manager
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

func NewOutPutGetter(ctx context.Context, repo templaterepo.TemplateRepo, m *managerparam.Manager) (Getter, error) {
	return &getter{
		templateRepo:       repo,
		templateMgr:        m.TemplateMgr,
		templateReleaseMgr: m.TemplateReleaseManager,
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

	chart, err := g.templateRepo.GetChart(tr.ChartName, releaseName, tr.LastSyncAt)
	if err != nil {
		return "", err
	}

	for _, file := range chart.Files {
		if file.Name == _outputsPath {
			return string(file.Data), nil
		}
	}

	return "", nil
}
