package manager

import (
	"context"
	"net/http"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/templaterelease/dao"
	"g.hz.netease.com/horizon/pkg/templaterelease/models"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
	"gorm.io/gorm"
)

const (
	_errCodeReleaseNotFound = errors.ErrorCode("ReleaseNotFound")
)

type Manager interface {
	// Create template release
	Create(ctx context.Context, templateRelease *models.TemplateRelease) (*models.TemplateRelease, error)
	// ListByTemplateName list all releases by template name
	ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error)
	// GetByTemplateNameAndRelease get release by template name and release name
	GetByTemplateNameAndRelease(ctx context.Context, templateName, release string) (*models.TemplateRelease, error)
}

func New(db *gorm.DB) Manager {
	return &manager{dao: dao.NewDAO(db)}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) Create(ctx context.Context,
	templateRelease *models.TemplateRelease) (*models.TemplateRelease, error) {
	return m.dao.Create(ctx, templateRelease)
}

func (m *manager) ListByTemplateName(ctx context.Context, templateName string) ([]*models.TemplateRelease, error) {
	return m.dao.ListByTemplateName(ctx, templateName)
}

func (m *manager) GetByTemplateNameAndRelease(ctx context.Context,
	templateName, release string) (_ *models.TemplateRelease, err error) {
	const op = "template release manager: get by template name and release"
	defer wlog.Start(ctx, op).StopPrint()

	tr, err := m.dao.GetByTemplateNameAndRelease(ctx, templateName, release)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			return nil, errors.E(op, http.StatusNotFound, _errCodeReleaseNotFound, err)
		}
		return nil, err
	}
	return tr, nil
}
