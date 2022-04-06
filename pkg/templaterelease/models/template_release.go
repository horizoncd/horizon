package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type TemplateRelease struct {
	global.Model

	TemplateName  string
	Name          string
	Description   string
	GitlabProject string
	Recommended   bool
	CreatedBy     uint
	UpdatedBy     uint
}
