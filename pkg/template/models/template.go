package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Template struct {
	global.Model

	Name        string
	ChartName   string
	Description string
	Token       string
	Repository  string
	GroupID     uint
	OnlyAdmin   *bool
	CreatedBy   uint
	UpdatedBy   uint
}
