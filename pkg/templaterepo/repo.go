package templaterepo

import (
	"fmt"
	"sync"
	"time"

	"helm.sh/helm/v3/pkg/chart"
)

const (
	cacheKeyFormat = "%s-%s"
)

//go:generate mockgen -source=$GOFILE -destination=../../mock/pkg/templaterepo/mock_repo.go -package=mock_repo
type TemplateRepo interface {
	GetLoc() string
	UploadChart(chart *chart.Chart) error
	DeleteChart(name string, version string) error
	ExistChart(name string, version string) (bool, error)
	GetChart(name string, version string, lastSyncAt time.Time) (*chart.Chart, error)
}

type RepoWithCache struct {
	TemplateRepo
	cache map[string]*ChartWithTime
	m     sync.Mutex
}

func NewRepoWithCache(repo TemplateRepo) TemplateRepo {
	return &RepoWithCache{
		TemplateRepo: repo,
		cache:        make(map[string]*ChartWithTime),
	}
}

type ChartWithTime struct {
	chartPkg   *chart.Chart
	lastSyncAt time.Time
}

func (r *RepoWithCache) GetLoc() string {
	return r.TemplateRepo.GetLoc()
}

func (r *RepoWithCache) UploadChart(chart *chart.Chart) error {
	return r.TemplateRepo.UploadChart(chart)
}

func (r *RepoWithCache) DeleteChart(name string, version string) error {
	return r.TemplateRepo.DeleteChart(name, version)
}

func (r *RepoWithCache) ExistChart(name string, version string) (bool, error) {
	return r.TemplateRepo.ExistChart(name, version)
}

func (r *RepoWithCache) GetChart(name string, version string, lastSyncAt time.Time) (*chart.Chart, error) {
	r.m.Lock()
	defer r.m.Unlock()
	cacheKey := fmt.Sprintf(cacheKeyFormat, name, version)
	if chartPkg, ok := r.cache[cacheKey]; ok && chartPkg != nil {
		if chartPkg.lastSyncAt.Sub(lastSyncAt) >= 0 {
			return chartPkg.chartPkg, nil
		}
	}

	chartPkg, err := r.TemplateRepo.GetChart(name, version, lastSyncAt)
	if err != nil {
		return nil, err
	}
	r.cache[cacheKey] = &ChartWithTime{
		chartPkg:   chartPkg,
		lastSyncAt: lastSyncAt,
	}
	return chartPkg, err
}
