package clean

import "time"

type WebhookLogCleanRule struct {
	TTL time.Duration `yaml:"ttl"`
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
