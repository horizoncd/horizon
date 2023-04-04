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

package webhook

import (
	"context"
	"testing"

	"github.com/horizoncd/horizon/pkg/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	utilcommon "github.com/horizoncd/horizon/pkg/util/common"
)

var (
	ctx = context.Background()
	db  *gorm.DB
	c   *controller

	resourceType     = "clusters"
	resourceID       = uint(1)
	updateWebhookReq = UpdateWebhookRequest{
		Enabled:          utilcommon.BoolPtr(true),
		URL:              utilcommon.StringPtr("http://xxxx"),
		SSLVerifyEnabled: utilcommon.BoolPtr(false),
		Triggers: []string{
			models.ClusterCreated,
		},
	}
	createWebhookReq = CreateWebhookRequest{
		URL:              "http://xxx",
		Enabled:          true,
		SSLVerifyEnabled: false,
		Triggers: []string{
			models.ClusterCreated,
		},
	}
)

func createContext() {
	db, _ = orm.NewSqliteDB("file::memory:?cache=shared")
	if err := db.AutoMigrate(
		&models.Webhook{},
		&models.WebhookLog{},
		&models.User{},
		&models.Event{},
		&models.Group{},
		&models.Application{},
		&models.Cluster{},
	); err != nil {
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
	mgrParam := managerparam.InitManager(db)
	controllerParam := param.Param{
		Manager: mgrParam,
	}
	c = NewController(&controllerParam).(*controller)
}

func Test(t *testing.T) {
	createContext()
	w, err := c.CreateWebhook(ctx, resourceType, resourceID, &createWebhookReq)
	assert.Nil(t, err)
	assert.Equal(t, createWebhookReq.URL, w.URL)

	w, err = c.GetWebhook(ctx, w.ID)
	assert.Nil(t, err)
	assert.Equal(t, createWebhookReq.URL, w.URL)

	uw := updateWebhookReq
	uw.URL = utilcommon.StringPtr("http://bbb")
	w, err = c.UpdateWebhook(ctx, w.ID, &uw)
	assert.Nil(t, err)
	assert.Equal(t, *(uw.URL), w.URL)

	ws, _, err := c.ListWebhooks(ctx, common.ResourceCluster, resourceID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ws))
	assert.Equal(t, *(uw.URL), w.URL)

	wl, err := c.webhookMgr.CreateWebhookLog(ctx, &models.WebhookLog{
		WebhookID: w.ID,
		URL:       w.URL,
		Status:    models.WebhookStatusWaiting,
	})
	assert.Nil(t, err)

	wlSummary, err := c.GetWebhookLog(ctx, wl.ID)
	assert.Nil(t, err)
	assert.Equal(t, wl.URL, wlSummary.URL)

	wl.Status = models.WebhookStatusSuccess
	wl, err = c.webhookMgr.UpdateWebhookLog(ctx, wl)
	assert.Nil(t, err)

	wlRetry, err := c.ResendWebhook(ctx, wl.ID)
	assert.Nil(t, err)
	assert.Equal(t, wl.URL, wlRetry.URL)
	assert.Equal(t, models.WebhookStatusWaiting, wlRetry.Status)

	query := q.New(nil)
	query.PageNumber = common.DefaultPageNumber
	query.PageSize = common.DefaultPageSize

	_, _, err = c.ListWebhookLogs(ctx, w.ID, query)
	assert.Nil(t, err)

	err = c.DeleteWebhook(ctx, w.ID)
	assert.Nil(t, err)

	event := models.Event{
		EventSummary: models.EventSummary{
			EventType: models.ApplicationCreated,
		},
	}

	ok, err := CheckIfEventMatch(&models.Webhook{
		Triggers: models.Any,
	}, &event)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)

	ok, err = CheckIfEventMatch(&models.Webhook{
		Triggers: models.ApplicationCreated,
	}, &event)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)

	ok, err = CheckIfEventMatch(&models.Webhook{
		Triggers: models.ClusterBuildDeployed,
	}, &event)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}
