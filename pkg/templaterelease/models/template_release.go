package models

import (
	"database/sql/driver"
	"time"

	"g.hz.netease.com/horizon/pkg/server/global"
)

type TemplateRelease struct {
	global.Model
	Template     uint
	TemplateName string
	ChartName    string

	// v1.0.0
	Name string
	Tag  string
	// v1.0.0-33da3204
	ChartVersion string

	Description  string
	Recommended  *bool
	OnlyAdmin    *bool
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
