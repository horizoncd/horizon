package manager

import (
	"context"
	"os"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/templateschematag/models"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

var (
	db  *gorm.DB
	ctx context.Context
)

func TestMain(m *testing.M) {
	db, _ = orm.NewSqliteDB("")
	if err := db.AutoMigrate(&models.ClusterTemplateSchemaTag{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	clusterID := uint(1)
	err := Mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
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

	tags, err := Mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "b", tags[1].Key)
	assert.Equal(t, "2", tags[1].Value)

	err = Mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
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

	tags, err = Mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "c", tags[1].Key)
	assert.Equal(t, "3", tags[1].Value)

	err = Mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
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

	tags, err = Mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 3, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "c", tags[1].Key)
	assert.Equal(t, "3", tags[1].Value)
	assert.Equal(t, "d", tags[2].Key)
	assert.Equal(t, "4", tags[2].Value)

	err = Mgr.UpsertByClusterID(ctx, clusterID, []*models.ClusterTemplateSchemaTag{
		{
			ClusterID: clusterID,
			Key:       "d",
			Value:     "4",
		},
	})
	assert.Nil(t, err)
	tags, err = Mgr.ListByClusterID(ctx, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "d", tags[0].Key)
	assert.Equal(t, "4", tags[0].Value)

	err = Mgr.UpsertByClusterID(ctx, clusterID, nil)
	assert.Nil(t, err)
	tags, err = Mgr.ListByClusterID(ctx, clusterID)
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
