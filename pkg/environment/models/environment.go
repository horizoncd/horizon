package models

import (
	"gorm.io/gorm"
	"gorm.io/plugin/soft_delete"
)

type Environment struct {
	gorm.Model

	Name        string
	DisplayName string
	CreatedBy   uint
	UpdatedBy   uint
	DeletedTs   soft_delete.DeletedAt
}

type EnvironmentList []*Environment

func (e EnvironmentList) Len() int {
	return len(e)
}

func (e EnvironmentList) Less(i, j int) bool {
	const pre = "pre"
	const online = "online"
	if e[i].Name == online {
		return false
	}
	if e[j].Name == online {
		return true
	}
	if e[i].Name == pre {
		return false
	}
	if e[j].Name == pre {
		return true
	}
	return e[i].Name < e[j].Name
}

func (e EnvironmentList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
