package models

import (
	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/pkg/server/global"
)

type CheckRunStatus string

const (
	CheckStatusPending    CheckRunStatus = "Pending"
	CheckStatusInProgress CheckRunStatus = "InProgress"
	CheckStatusSuccess    CheckRunStatus = "Success"
	CheckStatusFailure    CheckRunStatus = "Failure"
	CheckStatusCancelled  CheckRunStatus = "Cancelled"
)

var CheckRunStatusArr = [...]CheckRunStatus{
	CheckStatusPending,
	CheckStatusInProgress,
	CheckStatusSuccess,
	CheckStatusFailure,
	CheckStatusCancelled,
}

func String2CheckRunStatus(s string) CheckRunStatus {
	for _, status := range CheckRunStatusArr {
		if s == string(status) {
			return status
		}
	}
	return CheckStatusPending
}

type Check struct {
	global.Model
	common.Resource `json:",inline"`
}

type CheckRun struct {
	global.Model  `json:",inline"`
	Name          string         `json:"name"`
	CheckID       uint           `json:"checkId"`
	Status        CheckRunStatus `json:"status"`
	Message       string         `json:"message"`
	PipelineRunID uint           `gorm:"column:pipeline_run_id" json:"pipelineRunId"`
	DetailURL     string         `gorm:"column:detail_url" json:"detailUrl"`
}

func (CheckRun) TableName() string {
	return "tb_checkrun"
}
