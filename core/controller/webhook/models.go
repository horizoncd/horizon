package webhook

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"g.hz.netease.com/horizon/core/common"
	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/event/models"
	wmodels "g.hz.netease.com/horizon/pkg/webhook/models"
)

const (
	_triggerSeparator        = ","
	_resourceActionSeparator = "_"
)

type UpdateWebhookRequest struct {
	Enabled          *bool    `json:"enabled"`
	URL              *string  `json:"url"`
	SslVerifyEnabled *bool    `json:"sslVerifyEnabled"`
	Description      *string  `json:"description"`
	Secret           *string  `json:"secret"`
	Triggers         []string `json:"triggers"`
}

type CreateWebhookRequest struct {
	Enabled          bool     `json:"enabled"`
	URL              string   `json:"url"`
	SslVerifyEnabled bool     `json:"sslVerifyEnabled"`
	Description      string   `json:"description"`
	Secret           string   `json:"secret"`
	Triggers         []string `json:"triggers"`
	ResourceType     string   `json:"-"`
	ResourceID       uint     `json:"-"`
}

type Webhook struct {
	CreateWebhookRequest
	ID        uint         `json:"id"`
	CreatedAt time.Time    `json:"createdAt"`
	CreatedBy *common.User `json:"createdBy,omitempty"`
	UpdatedAt time.Time    `json:"updatedAt"`
	UpdatedBy *common.User `json:"updatedBy,omitempty"`
}

type LogSummary struct {
	ID           uint         `json:"id"`
	WebhookID    uint         `json:"webhookID"`
	EventID      uint         `json:"eventID"`
	URL          string       `json:"url"`
	Status       string       `json:"status"`
	ErrorMessage string       `json:"errorMessage"`
	CreatedAt    time.Time    `json:"createdAt"`
	CreatedBy    *common.User `json:"createdBy,omitempty"`
	UpdatedAt    time.Time    `json:"updatedAt"`
	UpdatedBy    *common.User `json:"updatedBy,omitempty"`
}

type Log struct {
	LogSummary
	RequestHeaders  string `json:"requestHeaders"`
	RequestData     string `json:"requestData"`
	ResponseHeaders string `json:"responseHeaders"`
	ResponseBody    string `json:"responseBody"`
}

func (w *UpdateWebhookRequest) toModel(ctx context.Context, wm *wmodels.Webhook) *wmodels.Webhook {
	if w.Enabled != nil {
		wm.Enabled = *w.Enabled
	}
	if w.URL != nil {
		wm.URL = *w.URL
	}
	if w.SslVerifyEnabled != nil {
		wm.SslVerifyEnabled = *w.SslVerifyEnabled
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

func (w *UpdateWebhookRequest) validate() error {
	if w.URL != nil {
		if err := validateURL(*w.URL); err != nil {
			return err
		}
	}
	if len(w.Triggers) > 0 {
		return validateTriggers(w.Triggers)
	}
	return nil
}

func (w *CreateWebhookRequest) toModel(ctx context.Context) *wmodels.Webhook {
	wm := &wmodels.Webhook{
		Enabled:          w.Enabled,
		URL:              w.URL,
		SslVerifyEnabled: w.SslVerifyEnabled,
		Description:      w.Description,
		Secret:           w.Secret,
		Triggers:         JoinTriggers(w.Triggers),
	}
	switch w.ResourceType {
	case common.ResourceGroup:
		wm.ResourceType = string(models.Group)
	case common.ResourceApplication:
		wm.ResourceType = string(models.Application)
	case common.ResourceCluster:
		wm.ResourceType = string(models.Cluster)
	}
	wm.ResourceID = w.ResourceID
	return wm
}

func (w *CreateWebhookRequest) validate() error {
	switch w.ResourceType {
	case common.ResourceGroup, common.ResourceApplication, common.ResourceCluster:
	default:
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid resource type: %s", w.ResourceType))
	}
	if err := validateURL(w.URL); err != nil {
		return err
	}
	if len(w.Triggers) <= 0 {
		return perror.Wrap(herrors.ErrParamInvalid, "triggers should not be empty")
	}
	return validateTriggers(w.Triggers)
}

func validateURL(url string) error {
	re := `^http(s)?://.+$`
	pattern := regexp.MustCompile(re)
	if !pattern.MatchString(url) {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid url, should satisfies the pattern %v", re))
	}
	return nil
}

func validateTriggers(triggers []string) error {
	validateClusterAction := func(action models.EventAction) bool {
		switch action {
		case models.Builded, models.Deployed, models.Freed, models.Rollbacked:
		default:
			return false
		}
		return true
	}

	validateApplicationAction := func(action models.EventAction) bool {
		switch action {
		case models.Transferred:
		default:
			return false
		}
		return true
	}

	validateCommonAction := func(action models.EventAction) bool {
		switch action {
		case models.AnyAction, models.Created, models.Deleted:
		default:
			return false
		}
		return true
	}

	for _, trigger := range triggers {
		resource, action, err := ParseResourceAction(trigger)
		if err != nil {
			return err
		}
		switch resource {
		case models.Cluster:
			if !validateCommonAction(action) && !validateClusterAction(action) {
				return perror.Wrap(herrors.ErrParamInvalid,
					fmt.Sprintf("invalid trigger: %s", trigger))
			}
		case models.Application:
			if !validateCommonAction(action) && !validateApplicationAction(action) {
				return perror.Wrap(herrors.ErrParamInvalid,
					fmt.Sprintf("invalid trigger: %s", trigger))
			}
		default:
			return perror.Wrap(herrors.ErrParamInvalid,
				fmt.Sprintf("invalid trigger: %s", trigger))
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

func ParseResourceAction(trigger string) (models.EventResourceType, models.EventAction, error) {
	parts := strings.Split(trigger, _resourceActionSeparator)
	if len(parts) != 2 {
		return "", "", perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid trigger %s", trigger))
	}
	resource := parts[0]
	action := parts[1]
	return models.EventResourceType(resource), models.EventAction(action), nil
}

func JoinResourceAction(resource, action string) string {
	return strings.Join([]string{resource, action}, _resourceActionSeparator)
}

func CheckIfEventMatch(webhook *wmodels.Webhook, event *models.Event) (bool, error) {
	triggers := ParseTriggerStr(webhook.Triggers)
	for _, trigger := range triggers {
		r, a, err := ParseResourceAction(trigger)
		if err != nil {
			return false, err
		}
		if (r == models.AnyResource || r == models.EventResourceType(event.ResourceType)) &&
			(a == models.AnyAction || a == models.EventAction(event.Action)) {
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
			SslVerifyEnabled: wm.SslVerifyEnabled,
			Description:      wm.Description,
			Secret:           wm.Secret,
			Triggers:         ParseTriggerStr(wm.Triggers),
			ResourceType:     wm.ResourceType,
			ResourceID:       wm.ResourceID,
		},
		ID:        wm.ID,
		CreatedAt: wm.CreatedAt,
		UpdatedAt: wm.UpdatedAt,
	}

	return w
}

func ofWebhookLogSummaryModel(wm *wmodels.WebhookLog) *LogSummary {
	wl := &LogSummary{
		ID:           wm.ID,
		WebhookID:    wm.WebhookID,
		EventID:      wm.EventID,
		URL:          wm.URL,
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
