package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

const (
	MessageTypeUser = iota
	MessageTypeSystem
)

type PRMessage struct {
	global.Model
	PipelineRunID uint `gorm:"column:pipeline_run_id"`
	Content       string
	MessageType   uint
	CreatedBy     uint
	UpdatedBy     uint
}

func (PRMessage) TableName() string {
	return "tb_pr_msg"
}
