package webhook

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/orm"
	"github.com/horizoncd/horizon/lib/q"
	applicationmodels "github.com/horizoncd/horizon/pkg/application/models"
	userauth "github.com/horizoncd/horizon/pkg/authentication/user"
	clustermodels "github.com/horizoncd/horizon/pkg/cluster/models"
	eventmodels "github.com/horizoncd/horizon/pkg/event/models"
	groupmodels "github.com/horizoncd/horizon/pkg/group/models"
	"github.com/horizoncd/horizon/pkg/param"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	utilcommon "github.com/horizoncd/horizon/pkg/util/common"
	webhookmodels "github.com/horizoncd/horizon/pkg/webhook/models"
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
			eventmodels.ClusterCreated,
		},
	}
	createWebhookReq = CreateWebhookRequest{
		URL:              "http://xxx",
		Enabled:          true,
		SSLVerifyEnabled: false,
		Triggers: []string{
			eventmodels.ClusterCreated,
		},
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

	wlRetry, err := c.ResendWebhook(ctx, wl.ID)
	assert.Nil(t, err)
	assert.Equal(t, wl.URL, wlRetry.URL)
	assert.Equal(t, webhookmodels.StatusWaiting, wlRetry.Status)

	query := q.New(nil)
	query.PageNumber = common.DefaultPageNumber
	query.PageSize = common.DefaultPageSize

	_, _, err = c.ListWebhookLogs(ctx, w.ID, query)
	assert.Nil(t, err)

	err = c.DeleteWebhook(ctx, w.ID)
	assert.Nil(t, err)
}
