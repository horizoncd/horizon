package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type TemplateRelease struct {
	global.Model
	Template     uint
	TemplateName string
	Name         string
	ChartName    string
	Description  string
	Recommended  *bool
	OnlyAdmin    *bool
	CreatedBy    uint
	UpdatedBy    uint
}
