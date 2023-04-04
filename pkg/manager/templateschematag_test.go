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
	"strings"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/models"
	"gorm.io/gorm"

	"github.com/stretchr/testify/assert"
)

func createTemplateSchemaTagCtx() (context.Context, *gorm.DB, TemplateSchemaTagManager) {
	var (
		db  *gorm.DB
		ctx context.Context
		mgr TemplateSchemaTagManager
	)

	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.ClusterTemplateSchemaTag{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()

	mgr = NewTemplateSchemaTagManager(db)

	return ctx, db, mgr
}

func TestTemplateSchemaTag(t *testing.T) {
	ctx, _, mgr := createTemplateSchemaTagCtx()
	clusterID := uint(1)
	err := mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
		{
			ClusterID: clusterID,
			Key:       "a",
			Value:     "1",
		}, {
			ClusterID: clusterID,
			Key:       "b",
			Value:     "2",
		},
	})
	assert.Nil(t, err)

	tags, err := mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "b", tags[1].Key)
	assert.Equal(t, "2", tags[1].Value)

	err = mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
		{
			ClusterID: clusterID,
			Key:       "a",
			Value:     "1",
		}, {
			ClusterID: clusterID,
			Key:       "c",
			Value:     "3",
		},
	})
	assert.Nil(t, err)

	tags, err = mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "c", tags[1].Key)
	assert.Equal(t, "3", tags[1].Value)

	err = mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
		{
			ClusterID: clusterID,
			Key:       "a",
			Value:     "1",
		}, {
			ClusterID: clusterID,
			Key:       "c",
			Value:     "3",
		}, {
			ClusterID: clusterID,
			Key:       "d",
			Value:     "4",
		},
	})
	assert.Nil(t, err)

	tags, err = mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 3, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "c", tags[1].Key)
	assert.Equal(t, "3", tags[1].Value)
	assert.Equal(t, "d", tags[2].Key)
	assert.Equal(t, "4", tags[2].Value)

	err = mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
		{
			ClusterID: clusterID,
			Key:       "d",
			Value:     "4",
		},
	})
	assert.Nil(t, err)
	tags, err = mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "d", tags[0].Key)
	assert.Equal(t, "4", tags[0].Value)

	err = mgr.UpsertByClusterID(ctx, clusterID, nil)
	assert.Nil(t, err)
	tags, err = mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, 0, len(tags))
}

func Test_ValidateUpsert(t *testing.T) {
	tags := make([]*models.ClusterTemplateSchemaTag, 0)
	tags = append(tags, &models.ClusterTemplateSchemaTag{
		Key:   strings.Repeat("a", 63),
		Value: "a",
	})
	err := ValidateUpsert(tags)
	assert.Nil(t, err)

	tags = tags[0:0]
	tags = append(tags, &models.ClusterTemplateSchemaTag{
		Key:   "",
		Value: "a",
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	tags = tags[0:0]
	tags = append(tags, &models.ClusterTemplateSchemaTag{
		Key:   "a",
		Value: "",
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	tags = append(tags, &models.ClusterTemplateSchemaTag{
		Key:   strings.Repeat("a", 64),
		Value: "a",
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	tags = tags[0:0]
	tags = append(tags, &models.ClusterTemplateSchemaTag{
		Key:   "a",
		Value: strings.Repeat("a", 1025),
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	tags = tags[0:0]
	tags = append(tags, &models.ClusterTemplateSchemaTag{
		Key:   "a(d",
		Value: "a",
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())
}
