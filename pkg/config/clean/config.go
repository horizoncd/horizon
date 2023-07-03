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

import "time"

type WebhookLogCleanRule struct {
	RelatedEventType string        `yaml:"relatedEventType"`
	TTL              time.Duration `yaml:"ttl"`
}

type EventCleanRule struct {
	EventType string        `yaml:"eventType"`
	TTL       time.Duration `yaml:"ttl"`

	// APIVersion only take effect when EventType is "clusters_kubernetes_event"
	APIVersion string `yaml:"apiVersion"`
	// Kind only take effect when EventType is "clusters_kubernetes_event"
	Kind string `yaml:"kind"`
	// Name only take effect when EventType is "clusters_kubernetes_event"
	Name string `yaml:"name"`
	// Namespace only take effect when EventType is "clusters_kubernetes_event"
	Namespace string `yaml:"namespace"`
	// Reason only take effect when EventType is "clusters_kubernetes_event"
	Reason string `yaml:"reason"`
}

type Config struct {
	Batch int `yaml:"batch"`
	// TimeToRun is a cron expression with seconds precision
	TimeToRun string `yaml:"timeToRun"`

	WebhookLogCleanRules []WebhookLogCleanRule `yaml:"webhookLogCleanRules"`
	EventCleanRules      []EventCleanRule      `yaml:"eventCleanRules"`
}
