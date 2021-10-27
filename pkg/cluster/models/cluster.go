package models

import "gorm.io/gorm"

type Cluster struct {
	gorm.Model

	Application         string
	Name                string
	Description         string
	GitURL              string
	GitSubfolder        string
	GitBranch           string
	Template            string
	TemplateRelease     string
	EnvironmentRegionID string
	CreatedBy           uint
	UpdatedBy           uint
}
