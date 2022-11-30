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
	ResourceType string       `json:"resourceType"`
	ResourceName string       `json:"resourceName"`
	ResourceID   uint         `json:"resourceID"`
	Action       string       `json:"action"`
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
		if err := validateURL(*w.URL); err != nil {
			return err
		}
	}
	if len(w.Triggers) > 0 {
		return c.validateTriggers(w.Triggers)
	}
	return nil
}

func (w *CreateWebhookRequest) toModel(ctx context.Context, resourceType string, resourceID uint) (*wmodels.Webhook, error) {
	wm := &wmodels.Webhook{
		Enabled:          w.Enabled,
		URL:              w.URL,
		SSLVerifyEnabled: w.SSLVerifyEnabled,
		Description:      w.Description,
		Secret:           w.Secret,
		Triggers:         JoinTriggers(w.Triggers),
	}
	switch resourceType {
	case common.ResourceGroup:
		wm.ResourceType = string(models.Group)
	case common.ResourceApplication:
		wm.ResourceType = string(models.Application)
	case common.ResourceCluster:
		wm.ResourceType = string(models.Cluster)
	default:
		return nil, perror.Wrap(herrors.ErrParamInvalid, fmt.Sprintf("invalid resource type: %s", resourceType))
	}
	wm.ResourceID = resourceID
	return wm, nil
}

func (c *controller) validateCreateRequest(resourceType string, w *CreateWebhookRequest) error {

	if err := c.validateResourceType(resourceType); err != nil {
		return err
	}
	if err := validateURL(w.URL); err != nil {
		return err
	}
	return c.validateTriggers(w.Triggers)
}
func validateURL(u string) error {
	re := `^http(s)?://.+$`
	pattern := regexp.MustCompile(re)
	if !pattern.MatchString(u) {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid url, should satisfies the pattern %v", re))
	}
	return nil
}

func (c *controller) validateResourceType(resource string) error {
	supportEvents := c.eventMgr.ListSupportEvents()
	if _, ok := supportEvents[models.EventResourceType(resource)]; !ok {
		return perror.Wrap(herrors.ErrParamInvalid,
			fmt.Sprintf("invalid resource type: %s", resource))
	}
	return nil
}

func (c *controller) validateTriggers(triggers []string) error {
	if len(triggers) <= 0 {
		return perror.Wrap(herrors.ErrParamInvalid, "triggers should not be empty")
	}
	// prepare supported events map
	supportEvents := map[models.EventResourceType]map[models.EventAction]bool{}
	for resource, actions := range c.eventMgr.ListSupportEvents() {
		for _, action := range actions {
			if _, ok := supportEvents[resource]; !ok {
				supportEvents[resource] = map[models.EventAction]bool{}
			}
			supportEvents[resource][action.Name] = true
		}
	}

	for _, trigger := range triggers {
		if trigger == models.Any {
			continue
		}
		resource, action, err := ParseResourceAction(trigger)
		if err != nil {
			return err
		}
		if _, ok := supportEvents[resource][action]; !ok {
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
		if trigger == models.Any {
			return true, nil
		}
		r, a, err := ParseResourceAction(trigger)
		if err != nil {
			return false, err
		}
		if (r == models.EventResourceType(event.ResourceType)) && (a == models.EventAction(event.Action)) {
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
		Action:       wm.Action,
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
