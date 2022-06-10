package manager

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"g.hz.netease.com/horizon/lib/orm"
	"g.hz.netease.com/horizon/pkg/applicationregion/models"

	"github.com/stretchr/testify/assert"
)

var (
	db, _ = orm.NewSqliteDB("")
	ctx   context.Context
	mgr   = New(db)
)

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&models.ApplicationRegion{}); err != nil {
		panic(err)
	}
	ctx = orm.NewContext(context.TODO(), db)
	os.Exit(m.Run())
}

func Test(t *testing.T) {
	applicationID := uint(1)
	err := mgr.UpsertByApplicationID(ctx, applicationID, []*models.ApplicationRegion{
		{
			ApplicationID:   applicationID,
			EnvironmentName: "test",
			RegionName:      "hz-test",
			CreatedBy:       1,
			UpdatedBy:       1,
		},
		{
			ApplicationID:   applicationID,
			EnvironmentName: "pre",
			RegionName:      "hz",
			CreatedBy:       1,
			UpdatedBy:       1,
		},
	})
	assert.Nil(t, err)

	appRegions, err := mgr.ListByApplicationID(ctx, applicationID)
	assert.Nil(t, err)
	assert.Equal(t, "pre", appRegions[0].EnvironmentName)
	assert.Equal(t, "hz", appRegions[0].RegionName)
	assert.Equal(t, uint(1), appRegions[0].UpdatedBy)
	for _, r := range appRegions {
		b, _ := json.Marshal(r)
		t.Logf("%v", string(b))
	}

	err = mgr.UpsertByApplicationID(ctx, applicationID, []*models.ApplicationRegion{
		{
			ApplicationID:   applicationID,
			EnvironmentName: "pre",
			RegionName:      "singapore",
			CreatedBy:       2,
			UpdatedBy:       2,
		},
	})
	assert.Nil(t, err)

	appRegions, err = mgr.ListByApplicationID(ctx, applicationID)
	assert.Nil(t, err)
	assert.Equal(t, "singapore", appRegions[0].RegionName)
	assert.Equal(t, uint(2), appRegions[0].UpdatedBy)
	for _, r := range appRegions {
		b, _ := json.Marshal(r)
		t.Logf("%v", string(b))
	}

	err = mgr.UpsertByApplicationID(ctx, applicationID, []*models.ApplicationRegion{
		{
			ApplicationID:   applicationID,
			EnvironmentName: "online",
			RegionName:      "singapore",
			CreatedBy:       2,
			UpdatedBy:       2,
		},
	})
	assert.Nil(t, err)

	appRegions, err = mgr.ListByApplicationID(ctx, applicationID)
	assert.Nil(t, err)
	assert.Equal(t, 3, len(appRegions))
	assert.Equal(t, "online", appRegions[0].EnvironmentName)
	assert.Equal(t, "singapore", appRegions[0].RegionName)
	assert.Equal(t, uint(2), appRegions[0].UpdatedBy)
	for _, r := range appRegions {
		b, _ := json.Marshal(r)
		t.Logf("%v", string(b))
	}
}
