package models

import "time"

const (
	StatusWaiting = "waiting"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

type Webhook struct {
	ID               uint
	Enabled          bool
	URL              string
	SSLVerifyEnabled bool
	Description      string
	Secret           string
	Triggers         string
	ResourceType     string
	ResourceID       uint
	CreatedAt        time.Time
	CreatedBy        uint
	UpdatedAt        time.Time
	UpdatedBy        uint
}

type WebhookLog struct {
	ID              uint
	WebhookID       uint
	EventID         uint
	URL             string
	RequestHeaders  string
	RequestData     string
	ResponseHeaders string
	ResponseBody    string
	Status          string
	ErrorMessage    string
	CreatedAt       time.Time
	CreatedBy       uint
	UpdatedAt       time.Time
}

type WebhookHistory struct {
	WebhookID    uint
	EventID      uint
	URL          string
	ResourceType string
	ResourceName string
	ResourceID   uint
	Action       string
	Status       string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type WebhookLogWithEventInfo struct {
	WebhookLog
	Action       string
	ResourceType string
	ResourceName string
	ResourceID   uint
}
