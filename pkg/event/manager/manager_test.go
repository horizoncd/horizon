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
	"testing"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	"github.com/stretchr/testify/assert"
)

var (
	ctx context.Context
	m   Manager
)

func createCtx() {
	db, _ := orm.NewSqliteDB("file::memory:?cache=shared")
	if err := db.AutoMigrate(&eventmodels.Event{},
		&eventmodels.EventCursor{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(); err != nil {
		panic(err)
	}
	ctx = context.Background()
	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: true,
	})
	m = New(db)
}

func Test(t *testing.T) {
	createCtx()
	e := &eventmodels.Event{
		EventSummary: eventmodels.EventSummary{
			ResourceType: common.ResourceCluster,
			ResourceID:   1,
			EventType:    eventmodels.ClusterCreated,
		},
	}
	e, err := m.CreateEvent(ctx, e)
	assert.Nil(t, err)

	e2, err := m.GetEvent(ctx, e.ID)
	assert.Nil(t, err)
	assert.Equal(t, e.ResourceID, e2.ResourceID)

	ec, err := m.CreateOrUpdateCursor(ctx, &eventmodels.EventCursor{
		Position: 1,
	})
	assert.Nil(t, err)

	ec.Position = 2
	_, err = m.CreateOrUpdateCursor(ctx, ec)
	assert.Nil(t, err)

	ec2, err := m.GetCursor(ctx)
	assert.Nil(t, err)
	assert.Equal(t, ec2.Position, ec2.Position)
}
