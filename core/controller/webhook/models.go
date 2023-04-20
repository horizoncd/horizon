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
	"fmt"
	"strings"
	"time"

	"github.com/horizoncd/horizon/core/common"
	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/event/models"
	usermodels "github.com/horizoncd/horizon/pkg/user/models"
	commonvalidate "github.com/horizoncd/horizon/pkg/util/validate"
	wmodels "github.com/horizoncd/horizon/pkg/webhook/models"
)

const (
	_triggerSeparator = ","
)

type UpdateWebhookRequest struct {
	Enabled          *bool    `json:"enabled"`
	URL              *string  `json:"url"`
	SSLVerifyEnabled *bool    `json:"sslVerifyEnabled"`
	Description      *string  `json:"description"`
	Secret           *string  `json:"secret"`
	Triggers         []string `json:"triggers"`
}

type CreateWebhookRequest struct {
	Enabled          bool     `json:"enabled"`
	URL              string   `json:"url"`
	SSLVerifyEnabled bool     `json:"sslVerifyEnabled"`
	Description      string   `json:"description"`
	Secret           string   `json:"secret"`
	Triggers         []string `json:"triggers"`
}

type Webhook struct {
	CreateWebhookRequest
	ID        uint                  `json:"id"`
	CreatedAt time.Time             `json:"createdAt"`
	CreatedBy *usermodels.UserBasic `json:"createdBy,omitempty"`
	UpdatedAt time.Time             `json:"updatedAt"`
	UpdatedBy *usermodels.UserBasic `json:"updatedBy,omitempty"`
}

type LogSummary struct {
	ID           uint                  `json:"id"`
	WebhookID    uint                  `json:"webhookID"`
	EventID      uint                  `json:"eventID"`
	URL          string                `json:"url"`
	Status       string                `json:"status"`
	ResourceType string                `json:"resourceType"`
	ResourceName string                `json:"resourceName"`
	ResourceID   uint                  `json:"resourceID"`
	EventType    string                `json:"eventType"`
	Extra        *string               `json:"extra"`
	ErrorMessage string                `json:"errorMessage"`
	CreatedAt    time.Time             `json:"createdAt"`
	CreatedBy    *usermodels.UserBasic `json:"createdBy,omitempty"`
	UpdatedAt    time.Time             `json:"updatedAt"`
	UpdatedBy    *usermodels.UserBasic `json:"updatedBy,omitempty"`
}

type Log struct {
	LogSummary
	RequestHeaders  string `json:"requestHeaders"`
	RequestData     string `json:"requestData"`
	ResponseHeaders string `json:"responseHeaders"`
	ResponseBody    string `json:"responseBody"`
}

func (w *UpdateWebhookRequest) toModel(wm *wmodels.Webhook) *wmodels.Webhook {
	if w.Enabled != nil {
		wm.Enabled = *w.Enabled
	}
	if w.URL != nil {
		wm.URL = *w.URL
	}
	if w.SSLVerifyEnabled != nil {
		wm.SSLVerifyEnabled = *w.SSLVerifyEnabled
	}
	if w.Description != nil {
		wm.Description = *w.Description
	}
	if w.Secret != nil {
		wm.Secret = *w.Secret
	}
	if len(w.Triggers) > 0 {
		wm.Triggers = JoinTriggers(w.Triggers)
	}
	return wm
}

func (c *controller) validateUpdateRequest(w *UpdateWebhookRequest) error {
	if w.URL != nil {
		if err := commonvalidate.CheckURL(*w.URL); err != nil {
			return err
		}
	}
	if len(w.Triggers) > 0 {
		return c.validateEvents(w.Triggers)
	}
	return nil
}

func (w *CreateWebhookRequest) toModel(ctx context.Context,
	resourceType string, resourceID uint) (*wmodels.Webhook, error) {
	wm := &wmodels.Webhook{
		ResourceType:     resourceType,
		ResourceID:       resourceID,
		Enabled:          w.Enabled,
		URL:              w.URL,
		SSLVerifyEnabled: w.SSLVerifyEnabled,
		Description:      w.Description,
		Secret:           w.Secret,
		Triggers:         JoinTriggers(w.Triggers),
	}
	return wm, nil
}

func (c *controller) validateCreateRequest(resourceType string, w *CreateWebhookRequest) error {
	if err := c.validateResourceType(resourceType); err != nil {
		return err
	}
	if err := commonvalidate.CheckURL(w.URL); err != nil {
		return err
	}

	if (!strings.HasPrefix(w.URL, "https")) && w.SSLVerifyEnabled {
		return perror.Wrapf(herrors.ErrParamInvalid, "sslVerifyEnabled is only valid for https")
	}

	return c.validateEvents(w.Triggers)
}

func (c *controller) validateResourceType(resource string) error {
	switch resource {
	case common.ResourceGroup, common.ResourceApplication, common.ResourceCluster:
	default:
		return perror.Wrapf(herrors.ErrParamInvalid, "invalid resource type %s", resource)
	}
	return nil
}

func (c *controller) validateEvents(events []string) error {
	if len(events) <= 0 {
		return perror.Wrap(herrors.ErrParamInvalid, "events should not be empty")
	}
	supportEvents := c.eventMgr.ListSupportEvents()

	for _, event := range events {
		if event == models.Any {
			continue
		}
		if _, ok := supportEvents[event]; !ok {
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("invalid event: %s", event))
		}
	}
	return nil
}

func ParseTriggerStr(triggerStr string) []string {
	return strings.Split(triggerStr, _triggerSeparator)
}

func JoinTriggers(triggers []string) string {
	return strings.Join(triggers, _triggerSeparator)
}

func CheckIfEventMatch(webhook *wmodels.Webhook, event *models.Event) (bool, error) {
	triggers := ParseTriggerStr(webhook.Triggers)
	for _, trigger := range triggers {
		if (trigger == models.Any) || (trigger == event.EventType) {
			return true, nil
		}
	}
	return false, nil
}

func ofWebhookModel(wm *wmodels.Webhook) *Webhook {
	w := &Webhook{
		CreateWebhookRequest: CreateWebhookRequest{
			Enabled:          wm.Enabled,
			URL:              wm.URL,
			SSLVerifyEnabled: wm.SSLVerifyEnabled,
			Description:      wm.Description,
			Secret:           wm.Secret,
			Triggers:         ParseTriggerStr(wm.Triggers),
		},
		ID:        wm.ID,
		CreatedAt: wm.CreatedAt,
		UpdatedAt: wm.UpdatedAt,
	}

	return w
}

func ofWebhookLogSummaryModel(wm *wmodels.WebhookLogWithEventInfo) *LogSummary {
	wl := &LogSummary{
		ID:           wm.ID,
		WebhookID:    wm.WebhookID,
		EventID:      wm.EventID,
		URL:          wm.URL,
		ResourceType: wm.ResourceType,
		ResourceID:   wm.ResourceID,
		ResourceName: wm.ResourceName,
		EventType:    wm.EventType,
		Status:       wm.Status,
		ErrorMessage: wm.ErrorMessage,
		CreatedAt:    wm.CreatedAt,
		UpdatedAt:    wm.UpdatedAt,
	}
	return wl
}

func ofWebhookLogModel(wm *wmodels.WebhookLog) *Log {
	wl := &Log{
		LogSummary: LogSummary{
			ID:           wm.ID,
			WebhookID:    wm.WebhookID,
			EventID:      wm.EventID,
			URL:          wm.URL,
			Status:       wm.Status,
			ErrorMessage: wm.ErrorMessage,
			CreatedAt:    wm.CreatedAt,
			UpdatedAt:    wm.UpdatedAt,
		},
		RequestHeaders:  wm.RequestHeaders,
		RequestData:     wm.RequestData,
		ResponseHeaders: wm.ResponseHeaders,
		ResponseBody:    wm.ResponseBody,
	}
	return wl
}
