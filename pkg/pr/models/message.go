package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

type PRMessage struct {
	global.Model
	PipelineRunID uint `gorm:"column:pipeline_run_id"`
	Content       string
	System        bool
	CreatedBy     uint
	UpdatedBy     uint
}

func (PRMessage) TableName() string {
	return "tb_pr_msg"
}
