// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package manager

import (
	"context"
	"regexp"

	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/pkg/dao"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"
)

type TemplateSchemaTagManager interface {
	// ListByClusterID Lists cluster tags by clusterID
	ListByClusterID(ctx context.Context, clusterID uint) ([]*models.ClusterTemplateSchemaTag, error)
	// UpsertByClusterID upsert cluster tags
	UpsertByClusterID(ctx context.Context, clusterID uint, tags []*models.ClusterTemplateSchemaTag) error
}

func NewTemplateSchemaTagManager(db *gorm.DB) TemplateSchemaTagManager {
	return &templateSchemaTagManager{
		dao: dao.NewTemplateSchemaTagDAO(db),
	}
}

type templateSchemaTagManager struct {
	dao dao.TemplateSchemaTagDAO
}

func (m *templateSchemaTagManager) ListByClusterID(ctx context.Context,
	clusterID uint) ([]*models.ClusterTemplateSchemaTag, error) {
	return m.dao.ListByClusterID(ctx, clusterID)
}

func (m *templateSchemaTagManager) UpsertByClusterID(ctx context.Context, clusterID uint,
	tags []*models.ClusterTemplateSchemaTag) error {
	return m.dao.UpsertByClusterID(ctx, clusterID, tags)
}

// ValidateUpsert tags upsert
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
