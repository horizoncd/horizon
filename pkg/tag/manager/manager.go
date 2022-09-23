package manager

import (
	"context"
	"regexp"
	"strings"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/tag/dao"
	"g.hz.netease.com/horizon/pkg/tag/models"
	"gorm.io/gorm"
)

type Manager interface {
	// ListByResourceTypeID ListByApplication tags by resourceType and resourceID
	ListByResourceTypeID(ctx context.Context, resourceType string, resourceID uint) ([]*models.Tag, error)
	// ListByResourceTypeIDs ListByApplication tags by resourceType and resourceID
	ListByResourceTypeIDs(ctx context.Context, resourceType string, resourceIDs []uint,
		deduplicate bool) ([]*models.Tag, error)
	// UpsertByResourceTypeID upsert tags
	UpsertByResourceTypeID(ctx context.Context, resourceType string, resourceID uint, tags []*models.Tag) error
}

func New(db *gorm.DB) Manager {
	return &manager{
		dao: dao.NewDAO(db),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) ListByResourceTypeID(ctx context.Context,
	resourceType string, resourceID uint) ([]*models.Tag, error) {
	return m.dao.ListByResourceTypeID(ctx, resourceType, resourceID)
}

func (m *manager) ListByResourceTypeIDs(ctx context.Context, resourceType string,
	resourceIDs []uint, deduplicate bool) ([]*models.Tag, error) {
	return m.dao.ListByResourceTypeIDs(ctx, resourceType, resourceIDs, deduplicate)
}

func (m *manager) UpsertByResourceTypeID(ctx context.Context,
	resourceType string, resourceID uint, tags []*models.Tag) error {
	return m.dao.UpsertByResourceTypeID(ctx, resourceType, resourceID, tags)
}

// ValidateUpsert tags upsert
func ValidateUpsert(tags []*models.Tag) error {
	if len(tags) > 20 {
		return perror.Wrap(herrors.ErrParamInvalid, "the count of tags must be less than 20")
	}
	keyPattern := regexp.MustCompile(`^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$`)
	for _, tag := range tags {
		if len(tag.Key) == 0 {
			return perror.Wrap(herrors.ErrParamInvalid, "tag key cannot be empty")
		}
		if len(tag.Key) > 63 {
			return perror.Wrapf(herrors.ErrParamInvalid, "tag key: %v is invalid, length must be 63 or less", tag.Key)
		}
		if len(tag.Value) > 1280 {
			return perror.Wrapf(herrors.ErrParamInvalid, "tag value: %v is invalid, length must be 1280 or less", tag.Value)
		}

		patternErr := perror.Wrapf(herrors.ErrParamInvalid, "tag key: %v is invalid, "+
			"should beginning and ending with an alphanumeric character ([a-z0-9A-Z]) "+
			"with dashes (-), slash(/), underscores (_), dots (.), and alphanumerics between", tag.Key)

		keySplit := strings.Split(tag.Key, "/")
		if len(keySplit) > 2 {
			return patternErr
		}
		for _, k := range keySplit {
			if k == "" || !keyPattern.MatchString(k) {
				return patternErr
			}
		}
	}
	return nil
}
