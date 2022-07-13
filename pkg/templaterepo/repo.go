package templaterepo

import "helm.sh/helm/v3/pkg/chart"

//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/templaterepo/mock_repo.go -package=mock_repo
type TemplateRepo interface {
	GetLoc() string
	UploadChart(chart *chart.Chart) error
	DeleteChart(name string, version string) error
	ExistChart(name string, version string) (bool, error)
	GetChart(name string, version string) (*chart.Chart, error)
}
