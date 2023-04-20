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

type WebhookLogWithEventInfo struct {
	WebhookLog
	EventType    string
	ResourceType string
	ResourceName string
	ResourceID   uint
	Extra        *string
}
