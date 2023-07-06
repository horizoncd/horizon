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
	"encoding/json"
	"time"

	"github.com/robfig/cron/v3"
	"k8s.io/apimachinery/pkg/util/runtime"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/lib/q"
	"github.com/horizoncd/horizon/pkg/config/clean"
	"github.com/horizoncd/horizon/pkg/event/models"
	"github.com/horizoncd/horizon/pkg/param/managerparam"
	"github.com/horizoncd/horizon/pkg/util/log"
)

const DefaultRule = ""

type Cleaner struct {
	clean.Config
	eventRules       map[string][]clean.EventCleanRule
	webhookRules     map[string][]clean.WebhookLogCleanRule
	mgr              *managerparam.Manager
	eventCursor      uint
	webhookLogCursor uint
}

func New(config clean.Config, mgr *managerparam.Manager) *Cleaner {
	if config.Batch == 0 {
		config.Batch = 160
	}
	eventCleanRules := make(map[string][]clean.EventCleanRule, len(config.EventCleanRules))
	webhookCleanRules := make(map[string][]clean.WebhookLogCleanRule, len(config.WebhookLogCleanRules))
	for _, rule := range config.EventCleanRules {
		eventCleanRules[rule.EventType] = append(eventCleanRules[rule.EventType], rule)
	}
	for _, rule := range config.WebhookLogCleanRules {
		webhookCleanRules[rule.RelatedEventType] = append(webhookCleanRules[rule.RelatedEventType], rule)
	}
	return &Cleaner{
		Config:       config,
		eventRules:   eventCleanRules,
		webhookRules: webhookCleanRules,
		mgr:          mgr,
	}
}

func (c *Cleaner) Run(ctx context.Context) {
	loc, err := time.LoadLocation("Asia/Shanghai")
	if err != nil {
		panic(err)
	}
	crontab := cron.New(cron.WithSeconds(), cron.WithLocation(loc))
	_, err = crontab.AddFunc(c.TimeToRun, func() {
		log.Info(ctx, "start to clean")
		current := time.Now()
		c.webhookLogClean(ctx, current)
		c.eventClean(ctx, current)
	})
	if err != nil {
		panic(err)
	}
	crontab.Run()
}

func (c *Cleaner) webhookLogClean(ctx context.Context, current time.Time) {
	defer runtime.HandleCrash()
	log.Debugf(ctx, "start to clean webhooklogs")
	defer func() {
		c.webhookLogCursor = 0
		log.Debugf(ctx, "finish to clean webhooklogs")
	}()
	if len(c.WebhookLogCleanRules) == 0 {
		return
	}
	needDeleted := make([]uint, 0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		query := &q.Query{
			Keywords: q.KeyWords{
				common.StartID: c.webhookLogCursor,
				common.Limit:   c.Batch,
				common.OrderBy: "l.id",
			},
		}
		webhooklogs, _, err := c.mgr.WebhookManager.ListWebhookLogs(ctx, query, nil)
		if err != nil {
			log.Errorf(ctx, "failed to list webhooklogs: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if len(webhooklogs) == 0 {
			return
		}

		maxID := uint(0)
		needDeleted = needDeleted[:0]
		for _, webhooklog := range webhooklogs {
			if webhooklog.ID > maxID {
				maxID = webhooklog.ID
			}
			rules := c.webhookRules[webhooklog.EventType]
			rules = append(rules, c.webhookRules[DefaultRule]...)
			for _, rule := range rules {
				if webhooklog.UpdatedAt.Add(rule.TTL).After(current) {
					continue
				}
				needDeleted = append(needDeleted, webhooklog.ID)
			}
		}

		c.webhookLogCursor = maxID
		if len(needDeleted) == 0 {
			continue
		}

		log.Infof(ctx, "need to delete webhooklogs: %v", needDeleted)
		_, _ = c.mgr.WebhookManager.DeleteWebhookLogs(ctx, needDeleted...)
	}
}

func (c *Cleaner) eventNeedClean(ctx context.Context, event *models.Event, current time.Time) bool {
	rules := c.eventRules[event.EventType]
	for _, rule := range rules {
		if current.Before(event.CreatedAt.Add(rule.TTL)) {
			continue
		}
		m := make(map[string]interface{})
		if event.Extra != nil && *event.Extra != "" {
			err := json.Unmarshal([]byte(*event.Extra), &m)
			if err != nil {
				log.Errorf(ctx, "failed to unmarshal event extra: %v", err)
				return true
			}
			if rule.Reason != "" && rule.Reason != m["reason"] {
				continue
			}
			involvedObject := m["involvedObject"].(map[string]interface{})
			if involvedObject != nil {
				if rule.APIVersion != "" && rule.APIVersion != involvedObject["apiVersion"] {
					continue
				}
				if rule.Kind != "" && rule.Kind != involvedObject["kind"] {
					continue
				}
				if rule.Name != "" && rule.Name != involvedObject["name"] {
					continue
				}
				if rule.Namespace != "" && rule.Namespace != involvedObject["namespace"] {
					continue
				}
			}
		}
		return true
	}
	return false
}

func (c *Cleaner) eventClean(ctx context.Context, current time.Time) {
	defer runtime.HandleCrash()
	log.Debugf(ctx, "start to clean events")
	defer func() {
		c.eventCursor = 0
		log.Debugf(ctx, "finish to clean events")
	}()
	if len(c.EventCleanRules) == 0 {
		return
	}
	needDeleted := make([]uint, 0)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}
		events, err := c.mgr.EventManager.ListEvents(ctx, &q.Query{Keywords: q.KeyWords{
			common.Limit:   c.Batch,
			common.StartID: int(c.eventCursor),
		}})
		if err != nil {
			log.Errorf(ctx, "failed to list events: %v", err)
			time.Sleep(5 * time.Second)
			continue
		}
		if len(events) == 0 {
			return
		}

		maxID := uint(0)
		needDeleted = needDeleted[:0]
		for _, event := range events {
			if event.ID > maxID {
				maxID = event.ID
			}
			if c.eventNeedClean(ctx, event, current) {
				needDeleted = append(needDeleted, event.ID)
			}
		}
		c.eventCursor = maxID
		if len(needDeleted) == 0 {
			continue
		}
		log.Infof(ctx, "need to delete event: %v", needDeleted)
		_, _ = c.mgr.EventManager.DeleteEvents(ctx, needDeleted...)
	}
}
