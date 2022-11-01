package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Template struct {
	global.Model

	Name        string
	ChartName   string
	Description string
	Repository  string
	GroupID     uint
	OnlyOwner   *bool
	WithoutCI   bool
	CreatedBy   uint
	UpdatedBy   uint
}
