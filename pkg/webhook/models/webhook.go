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
	SslVerifyEnabled bool
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
