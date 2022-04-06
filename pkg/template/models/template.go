package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Template struct {
	global.Model

	Name        string
	Description string
	CreatedBy   uint
	UpdatedBy   uint
}
