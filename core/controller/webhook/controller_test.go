package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/lib/orm"
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
	userauth "g.hz.netease.com/horizon/pkg/authentication/user"
	clustermodels "g.hz.netease.com/horizon/pkg/cluster/models"
	"g.hz.netease.com/horizon/pkg/event/models"
	eventmodels "g.hz.netease.com/horizon/pkg/event/models"
	groupmodels "g.hz.netease.com/horizon/pkg/group/models"
	"g.hz.netease.com/horizon/pkg/param"
	"g.hz.netease.com/horizon/pkg/param/managerparam"
	usermodels "g.hz.netease.com/horizon/pkg/user/models"
	utilcommon "g.hz.netease.com/horizon/pkg/util/common"
	webhookmodels "g.hz.netease.com/horizon/pkg/webhook/models"
)

var (
	ctx = context.Background()
	db  *gorm.DB
	c   *controller

	updateWebhookReq = UpdateWebhookRequest{
		Enabled:          utilcommon.BoolPtr(true),
		URL:              utilcommon.StringPtr("http://xxxx"),
		SslVerifyEnabled: utilcommon.BoolPtr(false),
		Triggers: []string{
			JoinResourceAction(string(eventmodels.Cluster), string(eventmodels.Created)),
		},
	}
	createWebhookReq = CreateWebhookRequest{
		URL:              "http://xxx",
		Enabled:          true,
		SslVerifyEnabled: false,
		Triggers: []string{
			JoinResourceAction(string(eventmodels.Cluster), string(eventmodels.Created)),
		},
		ResourceType: "clusters",
		ResourceID:   1,
	}
)

func createContext() {
	db, _ = orm.NewSqliteDB("file::memory:?cache=shared")
	if err := db.AutoMigrate(
		&webhookmodels.Webhook{},
		&webhookmodels.WebhookLog{},
		&usermodels.User{},
		&eventmodels.Event{},
		&groupmodels.Group{},
		&applicationmodels.Application{},
		&clustermodels.Cluster{},
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
	w, err := c.CreateWebhook(ctx, &createWebhookReq)
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

	ws, _, err := c.ListWebhooks(ctx, string(models.Cluster), createWebhookReq.ResourceID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(ws))
	assert.Equal(t, *(uw.URL), w.URL)

	wl, err := c.webhookMgr.CreateWebhookLog(ctx, &webhookmodels.WebhookLog{
		WebhookID: w.ID,
		URL:       w.URL,
		Status:    webhookmodels.StatusWaiting,
	})
	assert.Nil(t, err)

	wlSummary, err := c.GetWebhookLog(ctx, wl.ID)
	assert.Nil(t, err)
	assert.Equal(t, wl.URL, wlSummary.URL)

	wl.Status = webhookmodels.StatusSuccess
	wl, err = c.webhookMgr.UpdateWebhookLog(ctx, wl)
	assert.Nil(t, err)

	wlRetry, err := c.RetryWebhookLog(ctx, wl.ID)
	assert.Nil(t, err)
	assert.Equal(t, wl.URL, wlRetry.URL)
	assert.Equal(t, webhookmodels.StatusWaiting, wlRetry.Status)

	wlss, _, err := c.ListWebhookLogs(ctx, w.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 2, len(wlss))

	err = c.DeleteWebhook(ctx, w.ID)
	assert.Nil(t, err)

	wlss, _, err = c.ListWebhookLogs(ctx, w.ID, nil)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(wlss))
}
