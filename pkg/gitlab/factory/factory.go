package factory

import (
	"context"
	"fmt"
	"sync"

	herrors "github.com/horizoncd/horizon/core/errors"
	gitlablib "github.com/horizoncd/horizon/lib/gitlab"
	"github.com/horizoncd/horizon/pkg/config/gitlab"
	"github.com/horizoncd/horizon/pkg/util/wlog"
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
	defer wlog.Start(ctx, op).StopPrint()

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
		return nil, herrors.NewErrNotFound(herrors.GitlabResource,
			fmt.Sprintf("the gitlab instance for name: %s is not found.", name))
	}

	gitlabLib, err := gitlablib.New(gitlabModel.Token, gitlabModel.HTTPURL, gitlabModel.SSHURL)
	if err != nil {
		return nil, err
	}
	// store in cache
	f.m.Store(name, gitlabLib)
	return gitlabLib, nil
}
