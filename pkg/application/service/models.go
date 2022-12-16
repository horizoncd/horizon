package service

import (
	applicationmodels "github.com/horizoncd/horizon/pkg/application/models"
)

// ApplicationDetail contains the fullPath
type ApplicationDetail struct {
	applicationmodels.Application
	FullPath string
	FullName string
}
