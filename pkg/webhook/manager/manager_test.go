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
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	eventmanageer "github.com/horizoncd/horizon/pkg/event/manager"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	webhookmodels "github.com/horizoncd/horizon/pkg/webhook/models"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var (
	db, _    = orm.NewSqliteDB("")
	ctx      context.Context
	mgr      = New(db)
	eventMgr = eventmanageer.New(db)
)

func Test(t *testing.T) {
	clustersCreated := "clustersCreated"
	webhook := &webhookmodels.Webhook{
		Enabled:          true,
		URL:              "https://horizon.org",
		SSLVerifyEnabled: false,
		Triggers:         clustersCreated,
		ResourceType:     "clusters",
		ResourceID:       0,
	}

	webhook, err := mgr.CreateWebhook(ctx, webhook)
	assert.Nil(t, err)

	retrieveWebhook, err := mgr.GetWebhook(ctx, webhook.ID)
	assert.Nil(t, err)
	assert.NotNil(t, retrieveWebhook)
	assert.Equal(t, retrieveWebhook.ID, webhook.ID)

	resources := map[string][]uint{}
	resources[common.ResourceCluster] = []uint{0}
	_, count, err := mgr.ListWebhookOfResources(ctx, resources, q.New(q.KeyWords{
		common.Enabled: true,
	}))
	assert.Equal(t, int64(1), count)

	retrieveWebhooks, err := mgr.ListWebhooks(ctx)
	assert.Equal(t, 1, len(retrieveWebhooks))

	webhook.URL = "https://horizon.com"
	retrieveWebhook, err = mgr.UpdateWebhook(ctx, webhook.ID, webhook)
	assert.Equal(t, retrieveWebhook.ID, webhook.ID)
	assert.Equal(t, retrieveWebhook.URL, "https://horizon.com")

	events := []*eventmodels.Event{
		{
			EventSummary: eventmodels.EventSummary{
				ResourceType: common.ResourceCluster,
				ResourceID:   1,
				EventType:    eventmodels.ClusterCreated,
			},
			ReqID: "xxx",
		},
		{
			EventSummary: eventmodels.EventSummary{
				ResourceType: common.ResourceCluster,
				ResourceID:   2,
				EventType:    eventmodels.ClusterCreated,
			},
			ReqID: "xxx",
		},
	}
	for _, e := range events {
		_, err := eventMgr.CreateEvent(ctx, e)
		assert.Nil(t, err)
	}

	webhookLogs := []*webhookmodels.WebhookLog{
		{
			WebhookID:       1,
			EventID:         1,
			URL:             "http://example.com",
			RequestHeaders:  "Content-Type: application/json",
			RequestData:     `{"key": "value"}`,
			ResponseHeaders: "Content-Type: application/json",
			ResponseBody:    `{"status": "ok"}`,
			Status:          webhookmodels.StatusWaiting,
			ErrorMessage:    "",
		},
		{
			WebhookID:       2,
			EventID:         2,
			URL:             "http://example.com",
			RequestHeaders:  "Content-Type: application/json",
			RequestData:     `{"key": "value"}`,
			ResponseHeaders: "Content-Type: application/json",
			ResponseBody:    `{"status": "ok"}`,
			Status:          webhookmodels.StatusSuccess,
			ErrorMessage:    "",
		},
	}

	for _, log := range webhookLogs {
		_, err := mgr.CreateWebhookLog(ctx, log)
		assert.Nil(t, err)
	}

	retrievedLog, err := mgr.GetWebhookLog(ctx, webhookLogs[0].ID)
	assert.Nil(t, err)
	assert.NotNil(t, retrievedLog)
	assert.Equal(t, retrievedLog.ID, webhookLogs[0].ID)

	retrievedLog.Status = webhookmodels.StatusSuccess
	updatedLog, err := mgr.UpdateWebhookLog(ctx, retrievedLog)
	assert.Nil(t, err)
	assert.NotNil(t, updatedLog)
	assert.Equal(t, updatedLog.Status, webhookmodels.StatusSuccess)

	query := &q.Query{}
	logs, _, err := mgr.ListWebhookLogs(ctx, query, nil)
	assert.Nil(t, err)
	assert.NotNil(t, logs)
	assert.GreaterOrEqual(t, len(logs), 2)

	query = &q.Query{
		Keywords: q.KeyWords{
			common.StartID: 0,
			common.Limit:   10,
		},
	}
	cleanLogs, err := mgr.ListWebhookLogsForClean(ctx, query)
	assert.Nil(t, err)
	assert.NotNil(t, cleanLogs)
	assert.GreaterOrEqual(t, len(cleanLogs), 2)
	for _, log := range cleanLogs {
		assert.Contains(t, []uint{1, 2}, log.ID)
	}

	_, err = mgr.ResendWebhook(ctx, 1)
	assert.Nil(t, err)

	_, err = mgr.GetMaxEventIDOfLog(ctx)
	assert.Nil(t, err)

	retrievedLogs, err := mgr.GetWebhookLogByEventID(ctx, 1, 1)
	assert.Nil(t, err)
	assert.Equal(t, uint(1), retrievedLogs.ID)

	for _, log := range webhookLogs {
		_, err = mgr.DeleteWebhookLogs(ctx, log.ID)
		assert.Nil(t, err)
	}
}

func TestMain(m *testing.M) {
	if err := db.AutoMigrate(&webhookmodels.WebhookLog{},
		&webhookmodels.Webhook{}); err != nil {
		panic(err)
	}
	if err := db.AutoMigrate(&eventmodels.Event{}); err != nil {
		panic(err)
	}
	ctx = context.TODO()
	os.Exit(m.Run())
}
