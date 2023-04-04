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
	"encoding/json"
	"testing"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/models"

	"github.com/stretchr/testify/assert"
)

func TestApplicationRegion(t *testing.T) {
	var (
		db, _ = orm.NewSqliteDB("")
		ctx   context.Context
		mgr   = NewApplicationRegionManager(db)
	)
	if err := db.AutoMigrate(&models.ApplicationRegion{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()

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
