package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

type EnvironmentRegion struct {
	global.Model

	EnvironmentName string
	RegionName      string
	IsDefault       bool
	CreatedBy       uint
	UpdatedBy       uint
}
