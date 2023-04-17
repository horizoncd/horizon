package output

import (
	"errors"

	"golang.org/x/net/context"

	"github.com/horizoncd/horizon/pkg/param/managerparam"
	tmanager "github.com/horizoncd/horizon/pkg/template/manager"
	"github.com/horizoncd/horizon/pkg/templaterelease/manager"
	"github.com/horizoncd/horizon/pkg/templaterepo"
	"github.com/horizoncd/horizon/pkg/util/wlog"
)

// Getter provides some functions for output.
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
	// output yaml file path.
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
	templateName, releaseName string,
) (string, error) {
	const op = "template output getter: getTemplateOutPut"
	defer wlog.Start(ctx, op).StopPrint()

	tr, err := g.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return "", err
	}

	chart, err := g.templateRepo.GetChart(tr.ChartName, tr.ChartVersion, tr.LastSyncAt)
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
