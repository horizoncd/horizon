package gitlab

import (
	"context"
	"fmt"
	"sync"

	gitlablib "g.hz.netease.com/horizon/lib/gitlab"
	"g.hz.netease.com/horizon/pkg/gitlab"
	"g.hz.netease.com/horizon/util/errors"
	"g.hz.netease.com/horizon/util/wlog"
)

var (
	// Ctl Global instance of the gitlab controller
	Ctl = NewController()
)

type Controller interface {
	// GetByName get the gitlab.Instance by name
	GetByName(ctx context.Context, name string) (gitlablib.Interface, error)
}

type controller struct {
	// m use sync.Map for cache
	m *sync.Map
	// gitlabMgr to query gitlab db
	gitlabMgr gitlab.Manager
}

var _ Controller = (*controller)(nil)

// NewController initializes a new controller
func NewController() Controller {
	return &controller{
		m:         &sync.Map{},
		gitlabMgr: gitlab.Mgr,
	}
}

func (g *controller) GetByName(ctx context.Context, name string) (_ gitlablib.Interface, err error) {
	const op = "gitlab controller: get gitlab instance by name"
	defer wlog.Start(ctx, op).Stop(func() string { return wlog.ByErr(err) })

	var ret interface{}
	var ok bool
	// get from cache first
	if ret, ok = g.m.Load(name); ok {
		// exists in cache, return
		return ret.(gitlablib.Interface), nil
	}
	// not exists in cache, query db
	gitlabModel, err := g.gitlabMgr.GetByName(ctx, name)
	if err != nil {
		return nil, err
	}
	if gitlabModel == nil {
		// not in db
		errMsg := fmt.Sprintf("the gitlab instance for name: %s is not found. ", name)
		return nil, errors.E(op, errMsg)
	}

	gitlabLib, err := gitlablib.New(gitlabModel.Token, gitlabModel.URL)
	if err != nil {
		return nil, err
	}
	// store in cache
	g.m.Store(name, gitlabLib)
	return gitlabLib, nil
}
