package template

import (
	"context"
	"fmt"
	"net/http"

	"g.hz.netease.com/horizon/controller/gitlab"
	"g.hz.netease.com/horizon/pkg/templaterelease"
	"g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/wlog"
)

const (
	// schemaPath template input.schema.json file path
	schemaPath = "schema/input.schema.json"

	// ErrCodeReleaseNotFound  ReleaseNotFound error code
	ErrCodeReleaseNotFound = errors.ErrorCode("ReleaseNotFound")
)

var (
	Ctl = NewController()
)

type Controller interface {
	// GetTemplateSchema get schema for a template release
	GetTemplateSchema(ctx context.Context, templateName, releaseName string) (_ []byte, err error)
}

type controller struct {
	templateReleaseMgr templaterelease.Manager
	gitlabCtl          gitlab.Controller
}

var _ Controller = (*controller)(nil)

// NewController initializes a new controller
func NewController() Controller {
	return &controller{
		templateReleaseMgr: templaterelease.Mgr,
		gitlabCtl:          gitlab.Ctl,
	}
}

func (c *controller) GetTemplateSchema(ctx context.Context, templateName, releaseName string) (_ []byte, err error) {
	const op = "template controller: getTemplateSchema"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	tr, err := c.templateReleaseMgr.GetByTemplateNameAndRelease(ctx, templateName, releaseName)
	if err != nil {
		return nil, err
	}
	if tr == nil {
		return nil, errors.E(op, http.StatusNotFound, ErrCodeReleaseNotFound,
			fmt.Sprintf("the release %v of template %v is not found", releaseName, templateName))
	}
	gitlabLib, err := c.gitlabCtl.GetByName(ctx, tr.GitlabName)
	if err != nil {
		return nil, err
	}
	b, err := gitlabLib.GetFile(ctx, tr.GitlabProject, tr.Name, schemaPath)
	if err != nil {
		return nil, err
	}
	return b, nil
}
