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

package clean

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/pkg/config/clean"
	"github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	wmodels "github.com/horizoncd/horizon/pkg/webhook/models"
)

func TestEventClean(t *testing.T) {
	db, err := orm.NewSqliteDB("")
	assert.Nil(t, err)
	err = db.AutoMigrate(&models.Event{}, &wmodels.WebhookLog{}, &wmodels.Webhook{})
	assert.Nil(t, err)

	ctx := context.TODO()
	mgr := managerparam.InitManager(db)

	clustersCreated := "clustersCreated"
	clustersKubernetesEvent := "clustersKubernetesEvent"

	webhook := &wmodels.Webhook{
		Enabled:          true,
		URL:              "https://horizon.org",
		SSLVerifyEnabled: false,
		Triggers:         clustersCreated,
		ResourceType:     "clusters",
		ResourceID:       0,
	}

	webhook, err = mgr.WebhookMgr.CreateWebhook(ctx, webhook)
	assert.Nil(t, err)

	durationKept := time.Duration(0)
	durationDeleted := time.Hour * 28
	durationThreshold := time.Hour * 24

	event := &models.Event{
		EventSummary: models.EventSummary{
			ResourceType: "clusters",
			ResourceID:   0,
			EventType:    clustersCreated,
		},
		ReqID:     "xxx",
		CreatedAt: time.Now().Add(-durationKept),
	}

	events, err := mgr.EventMgr.CreateEvent(ctx, event)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(events))
	eventNeedToKept := events[0]

	webhookNeedToDelete := &wmodels.WebhookLog{
		WebhookID: webhook.ID,
		EventID:   event.ID,
		URL:       webhook.URL,
		CreatedAt: time.Now().Add(-durationDeleted),
		UpdatedAt: time.Now().Add(-durationDeleted),
	}
	webhookNeedToDelete, err = mgr.WebhookMgr.CreateWebhookLog(ctx, webhookNeedToDelete)
	assert.Nil(t, err)

	webhookNeedToKeep := &wmodels.WebhookLog{
		WebhookID: webhook.ID,
		EventID:   event.ID,
		URL:       webhook.URL,
		CreatedAt: time.Now().Add(-durationKept),
		UpdatedAt: time.Now().Add(-durationKept),
	}
	webhookNeedToKeep, err = mgr.WebhookMgr.CreateWebhookLog(ctx, webhookNeedToKeep)
	assert.Nil(t, err)

	event = &models.Event{
		EventSummary: models.EventSummary{
			ResourceType: "clusters",
			ResourceID:   0,
			EventType:    clustersCreated,
		},
		ReqID:     "xxx",
		CreatedAt: time.Now().Add(-durationDeleted),
	}

	events, err = mgr.EventMgr.CreateEvent(ctx, event)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(events))
	eventNeedToDelete := events[0]

	extra := `{"involvedObject":{"apiVersion":"argoproj.io/v1alpha1",` +
		`"kind":"Application","name":"yufeng-helloworld-serverless-jjy","namespace":"argocd"},` +
		`"lastTimestamp":"2023-06-14T03:18:21Z",` +
		`"message":"Updated health status: Healthy -\u003e Unknown",` +
		`"name":"yufeng-helloworld-serverless-jjy.17686843f676d6f2",` +
		`"reason":"ResourceUpdated","type":"Normal"}`
	event = &models.Event{
		EventSummary: models.EventSummary{
			ResourceType: "clusters",
			ResourceID:   0,
			EventType:    clustersKubernetesEvent,
			Extra:        &extra,
		},
		ReqID:     "xxx",
		CreatedAt: time.Now().Add(-durationDeleted),
	}

	events, err = mgr.EventMgr.CreateEvent(ctx, event)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(events))
	eventNeedToDelete2 := events[0]

	cleaner := New(clean.Config{
		Batch:                2,
		WebhookLogCleanRules: []clean.WebhookLogCleanRule{{RelatedEventType: clustersCreated, TTL: durationThreshold}},
		EventCleanRules: []clean.EventCleanRule{
			{EventType: clustersCreated, TTL: durationThreshold},
			{
				EventType:  clustersKubernetesEvent,
				TTL:        durationThreshold,
				APIVersion: "argoproj.io/v1alpha1",
				Kind:       "Application",
				Name:       "yufeng-helloworld-serverless-jjy",
				Namespace:  "argocd",
				Reason:     "ResourceUpdated",
			},
		},
	}, mgr)

	current := time.Now()
	cleaner.webhookLogClean(ctx, current)
	cleaner.eventClean(ctx, current)

	_, err = mgr.EventMgr.GetEvent(ctx, eventNeedToKept.ID)
	assert.Nil(t, err)

	_, err = mgr.EventMgr.GetEvent(ctx, eventNeedToDelete.ID)
	assert.NotNil(t, err)

	_, err = mgr.EventMgr.GetEvent(ctx, eventNeedToDelete2.ID)
	assert.NotNil(t, err)

	_, err = mgr.WebhookMgr.GetWebhookLog(ctx, webhookNeedToKeep.ID)
	assert.Nil(t, err)

	_, err = mgr.WebhookMgr.GetWebhookLog(ctx, webhookNeedToDelete.ID)
	assert.NotNil(t, err)
}
