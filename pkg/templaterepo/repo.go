package templaterepo

import (
	"fmt"
	"sync"
	"time"

	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/pkg/config/templaterepo"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"helm.sh/helm/v3/pkg/chart"
)

const (
	cacheKeyFormat = "%s-%s"
)

type Constructor func(repo templaterepo.Repo) (TemplateRepo, error)

var factory = make(map[string]Constructor)

func Register(tp string, constructor Constructor) {
	factory[tp] = constructor
}

func NewRepo(config templaterepo.Repo) (TemplateRepo, error) {
	if constructor, ok := factory[config.Kind]; ok {
		repo, err := constructor(config)
		if err != nil {
			return nil, err
		}
		return NewRepoWithCache(repo), nil
	}
	return nil, perror.Wrapf(herrors.ErrParamInvalid, "type (%s) not implement", config.Kind)
}

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
