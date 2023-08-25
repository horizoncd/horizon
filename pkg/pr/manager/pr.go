package manager

import (
	"gorm.io/gorm"
)

type PRManager struct {
	PipelineRun PipelineRunManager
	Message     PRMessageManager
	Check       CheckManager
}

func NewPRManager(db *gorm.DB) *PRManager {
	return &PRManager{
		PipelineRun: NewPipelineRunManager(db),
		Message:     NewPRMessageManager(db),
		Check:       NewCheckManager(db),
	}
}
