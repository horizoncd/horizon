package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

type Template struct {
	global.Model

	Name string
	// TODO: remove the ChartName, now ChartName equals Name
	ChartName   string
	Description string
	Repository  string
	GroupID     uint
	OnlyOwner   *bool
	WithoutCI   bool
	CreatedBy   uint
	UpdatedBy   uint
}
