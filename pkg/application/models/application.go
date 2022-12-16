package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
)

type Priority string

const (
	P0 Priority = "P0"
	P1 Priority = "P1"
	P2 Priority = "P2"
	P3 Priority = "P3"
)

type Application struct {
	global.Model
	GroupID         uint
	Name            string
	Description     string
	Priority        Priority
	GitURL          string
	GitSubfolder    string
	GitRef          string
	GitRefType      string
	Template        string
	TemplateRelease string
	CreatedBy       uint
	UpdatedBy       uint
}
