package manager

import (
	"context"
	"fmt"
	"regexp"

	"g.hz.netease.com/horizon/pkg/clustertag/dao"
	"g.hz.netease.com/horizon/pkg/clustertag/models"
)

var (
	// Mgr is the global cluster tag manager
	Mgr = New()
)

type Manager interface {
	// ListByClusterID List cluster tags by clusterID
	ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTag, error)
	// UpsertByClusterID upsert cluster tags
	UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTag) error
}

func New() Manager {
	return &manager{
		dao: dao.NewDAO(),
	}
}

type manager struct {
	dao dao.DAO
}

func (m *manager) ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTag, error) {
	return m.dao.ListByClusterID(ctx, clusterID)
}

func (m *manager) UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTag) error {
	return m.dao.UpsertByClusterID(ctx, clusterID, tags)
}

// ValidateUpsert tags upsert
func ValidateUpsert(tags []*models.ClusterTag) error {
	if len(tags) > 20 {
		return fmt.Errorf("the count of tags must be less than 20")
	}
	keyPattern := regexp.MustCompile(`^(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])?$`)
	for _, tag := range tags {
		if len(tag.Key) == 0 {
			return fmt.Errorf("tag key cannot be empty")
		}
		if len(tag.Value) == 0 {
			return fmt.Errorf("tag value cannot be empty")
		}
		if len(tag.Key) > 63 {
			return fmt.Errorf("tag key: %v is invalid, length must be 63 or less", tag.Key)
		}
		if len(tag.Value) > 1024 {
			return fmt.Errorf("tag value: %v is invalid, length must be 1024 or less", tag.Value)
		}

		if !keyPattern.MatchString(tag.Key) {
			return fmt.Errorf("tag key: %v is invalid, "+
				"should beginning and ending with an alphanumeric character ([a-z0-9A-Z]) "+
				"with dashes (-), underscores (_), dots (.), and alphanumerics between", tag.Key)
		}
	}
	return nil
}
