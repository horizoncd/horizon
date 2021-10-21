package factory

import (
	"context"
	"fmt"
	"sync"

	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	"g.hz.netease.com/horizon/pkg/util/errors"
	"g.hz.netease.com/horizon/pkg/util/wlog"
)

// Factory used to get the gitlab instance
type Factory interface {
	// GetByName get the gitlablib.Interface by name
	GetByName(ctx context.Context, name string) (gitlablib.Interface, error)
}

type factory struct {
	// m use sync.Map for cache
	m            *sync.Map
	gitlabMapper gitlab.Mapper
}

var _ Factory = (*factory)(nil)

// NewFactory initializes a new factory
func NewFactory(gitlabMapper gitlab.Mapper) Factory {
	return &factory{
		m:            &sync.Map{},
		gitlabMapper: gitlabMapper,
	}
}

func (f *factory) GetByName(ctx context.Context, name string) (_ gitlablib.Interface, err error) {
	const op = "gitlab controller: get gitlab instance by name"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	var ret interface{}
	var ok bool
	// get from cache first
	if ret, ok = f.m.Load(name); ok {
		// exists in cache, return
		return ret.(gitlablib.Interface), nil
	}
	// not exists in cache
	gitlabModel, ok := f.gitlabMapper[name]
	if !ok {
		errMsg := fmt.Sprintf("the gitlab instance for name: %s is not found. ", name)
		return nil, errors.E(op, errMsg)
	}

	gitlabLib, err := gitlablib.New(gitlabModel.Token, gitlabModel.HTTPURL)
	if err != nil {
		return nil, err
	}
	// store in cache
	f.m.Store(name, gitlabLib)
	return gitlabLib, nil
}
