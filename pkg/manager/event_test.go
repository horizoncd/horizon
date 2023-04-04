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
	"github.com/horizoncd/horizon/lib/q"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/models"
	"github.com/stretchr/testify/assert"
)

func createCtx() (context.Context, EventManager) {
	db, _ := orm.NewSqliteDB("file::memory:?cache=shared")
	if err := db.AutoMigrate(&models.Event{},
		&models.EventCursor{}, &models.WebhookLog{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(); err != nil {
		panic(err)
	}
	ctx := context.Background()
	// nolint
	ctx = common.WithContext(ctx, &userauth.DefaultInfo{
		Name:  "Jerry",
		ID:    1,
		Admin: true,
	})
	m := NewEventManager(db)
	return ctx, m
}

func TestEvent(t *testing.T) {
	ctx, m := createCtx()
	e := &models.Event{
		EventSummary: models.EventSummary{
			ResourceType: common.ResourceCluster,
			ResourceID:   1,
			EventType:    models.ClusterCreated,
		},
		ReqID: "xxx",
	}
	events, err := m.CreateEvent(ctx, e)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(events))

	e = events[0]
	e2, err := m.GetEvent(ctx, e.ID)
	assert.Nil(t, err)
	assert.Equal(t, e.ResourceID, e2.ResourceID)

	events, err = m.ListEvents(ctx, &q.Query{Keywords: q.KeyWords{common.ReqID: "xxx"}})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(events))

	e.ID = 3
	events, err = m.CreateEvent(ctx, e)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(events))

	events, err = m.ListEvents(ctx, &q.Query{Keywords: q.KeyWords{common.ReqID: "xxx"}})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(events))

	events, err = m.ListEvents(ctx, &q.Query{Keywords: q.KeyWords{common.StartID: 2, common.Limit: 10}})
	assert.Nil(t, err)
	assert.Equal(t, 1, len(events))

	events, err = m.ListEvents(ctx, &q.Query{Keywords: q.KeyWords{common.StartID: 0, common.Limit: 10}})
	assert.Nil(t, err)
	assert.Equal(t, 2, len(events))

	ec, err := m.CreateOrUpdateCursor(ctx, &models.EventCursor{
		Position: 1,
	})
	assert.Nil(t, err)

	ec.Position = 2
	_, err = m.CreateOrUpdateCursor(ctx, ec)
	assert.Nil(t, err)

	ec2, err := m.GetCursor(ctx)
	assert.Nil(t, err)
	assert.Equal(t, ec2.Position, ec2.Position)

	_, err = m.DeleteEvents(ctx, 1, 3)
	assert.Nil(t, err)

	events, err = m.ListEvents(ctx, &q.Query{Keywords: q.KeyWords{common.StartID: 0, common.Limit: 10}})
	assert.Nil(t, err)
	assert.Equal(t, 0, len(events))
}
