package manager

import (
	"context"
	"regexp"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/templateschematag/dao"
	"github.com/horizoncd/horizon/pkg/templateschematag/models"
	"gorm.io/gorm"
)

type Manager interface {
	// ListByClusterID Lists cluster tags by clusterID
	ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTemplateSchemaTag, error)
	// UpsertByClusterID upsert cluster tags
	UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTemplateSchemaTag) error
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTemplateSchemaTag, error) {
	return m.dao.ListByClusterID(ctx, clusterID)
}

func (m *manager) UpsertByClusterID(ctx context.Context, clusterID uint,
	tags []*models.ClusterTemplateSchemaTag,
) error {
	return m.dao.UpsertByClusterID(ctx, clusterID, tags)
}

// ValidateUpsert tags upsert.
func ValidateUpsert(tags []*models.ClusterTemplateSchemaTag) error {
	if len(tags) > 20 {
		return perror.WithMessage(herrors.ErrParamInvalid, "the count of tags must be less than 20")
	}
	keyPattern := regexp.MustCompile(`^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$`)
	for _, tag := range tags {
		if len(tag.Key) == 0 {
			return perror.WithMessage(herrors.ErrParamInvalid, "tag key cannot be empty")
		}
		if len(tag.Value) == 0 {
			return perror.WithMessage(herrors.ErrParamInvalid, "tag value cannot be empty")
		}
		if len(tag.Key) > 63 {
			return perror.WithMessagef(herrors.ErrParamInvalid,
				"tag key: %v is invalid, length must be 63 or less", tag.Key)
		}
		if len(tag.Value) > 1024 {
			return perror.WithMessagef(herrors.ErrParamInvalid,
				"tag value: %v is invalid, length must be 1024 or less", tag.Value)
		}

		if !keyPattern.MatchString(tag.Key) {
			return perror.WithMessagef(herrors.ErrParamInvalid,
				"tag key: %v is invalid, should beginning and ending "+
					"with an alphanumeric character ([a-z0-9A-Z]) "+
					"with dashes (-), underscores (_), dots (.), and alphanumerics between", tag.Key)
		}
	}
	return nil
}
