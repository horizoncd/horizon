package service

import (
	applicationmodels "g.hz.netease.com/horizon/pkg/application/models"
)

// ApplicationDetail contains the fullPath
type ApplicationDetail struct {
	applicationmodels.Application
	FullPath string
}
