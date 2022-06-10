package manager

import (
	"context"
	"os"
	"strings"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/tag/models"

	"github.com/stretchr/testify/assert"
)

var (
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.Tag{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	clusterID := uint(1)
	err := mgr.UpsertByResourceTypeID(ctx, models.TypeCluster, clusterID, []*models.Tag{
		{
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "a",
			Value:        "1",
		}, {
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "b",
			Value:        "2",
		},
	})
	assert.Nil(t, err)

	tags, err := mgr.ListByResourceTypeID(ctx, models.TypeCluster, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "b", tags[1].Key)
	assert.Equal(t, "2", tags[1].Value)

	err = mgr.UpsertByResourceTypeID(ctx, models.TypeCluster, clusterID, []*models.Tag{
		{
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "a",
			Value:        "1",
		}, {
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "c",
			Value:        "3",
		},
	})
	assert.Nil(t, err)

	tags, err = mgr.ListByResourceTypeID(ctx, models.TypeCluster, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 2, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "c", tags[1].Key)
	assert.Equal(t, "3", tags[1].Value)

	err = mgr.UpsertByResourceTypeID(ctx, models.TypeCluster, clusterID, []*models.Tag{
		{
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "a",
			Value:        "1",
		}, {
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "c",
			Value:        "3",
		}, {
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "d",
			Value:        "4",
		},
	})
	assert.Nil(t, err)

	tags, err = mgr.ListByResourceTypeID(ctx, models.TypeCluster, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 3, len(tags))
	assert.Equal(t, "a", tags[0].Key)
	assert.Equal(t, "1", tags[0].Value)
	assert.Equal(t, "c", tags[1].Key)
	assert.Equal(t, "3", tags[1].Value)
	assert.Equal(t, "d", tags[2].Key)
	assert.Equal(t, "4", tags[2].Value)

	err = mgr.UpsertByResourceTypeID(ctx, models.TypeCluster, clusterID, []*models.Tag{
		{
			ResourceType: models.TypeCluster,
			ResourceID:   clusterID,
			Key:          "d",
			Value:        "4",
		},
	})
	assert.Nil(t, err)
	tags, err = mgr.ListByResourceTypeID(ctx, models.TypeCluster, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, tags)
	assert.Equal(t, 1, len(tags))
	assert.Equal(t, "d", tags[0].Key)
	assert.Equal(t, "4", tags[0].Value)

	err = mgr.UpsertByResourceTypeID(ctx, models.TypeCluster, clusterID, nil)
	assert.Nil(t, err)
	tags, err = mgr.ListByResourceTypeID(ctx, models.TypeCluster, clusterID)
	assert.Nil(t, err)
	assert.NotNil(t, 0, len(tags))
}

func Test_ValidateUpsert(t *testing.T) {
	tags := make([]*models.Tag, 0)
	tags = append(tags, &models.Tag{
		Key:   strings.Repeat("a", 63),
		Value: "a",
	})
	err := ValidateUpsert(tags)
	assert.Nil(t, err)

	tags = tags[0:0]
	tags = append(tags, &models.Tag{
		Key:   "",
		Value: "a",
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	tags = append(tags, &models.Tag{
		Key:   strings.Repeat("a", 64),
		Value: "a",
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	tags = tags[0:0]
	tags = append(tags, &models.Tag{
		Key:   "a",
		Value: strings.Repeat("a", 1281),
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())

	tags = tags[0:0]
	tags = append(tags, &models.Tag{
		Key:   "a(d",
		Value: "a",
	})
	err = ValidateUpsert(tags)
	assert.NotNil(t, err)
	t.Logf("%v", err.Error())
}
