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

import (
	"database/sql/driver"
	"time"

	"github.com/horizoncd/horizon/pkg/server/global"
)

type TemplateRelease struct {
	global.Model
	Template     uint
	TemplateName string
	ChartName    string

	// v1.0.0
	Name string
	// v1.0.0-33da3204
	ChartVersion string

	Description  string
	Recommended  *bool
	OnlyOwner    *bool
	SyncStatus   SyncStatus
	LastSyncAt   time.Time
	FailedReason string
	CommitID     string
	CreatedBy    uint
	UpdatedBy    uint
}

type SyncStatus uint8

const (
	StatusSucceed SyncStatus = iota + 1
	StatusUnknown
	StatusFailed
	StatusOutOfSync
)

func (s *SyncStatus) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		*s = 0
	}
	str := string(bytes)

	switch str {
	case "status_succeed":
		*s = StatusSucceed
	case "status_unknown":
		*s = StatusUnknown
	case "status_failed":
		*s = StatusFailed
	case "status_outofsync":
		*s = StatusOutOfSync
	}
	return nil
}

func (s SyncStatus) Value() (driver.Value, error) {
	if s > StatusOutOfSync || s < StatusSucceed {
		return nil, nil
	}
	switch s {
	case StatusSucceed:
		return "status_succeed", nil
	case StatusUnknown:
		return "status_unknown", nil
	case StatusFailed:
		return "status_failed", nil
	case StatusOutOfSync:
		return "status_outofsync", nil
	default:
		return nil, nil
	}
}
