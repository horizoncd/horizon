package code

import (
	"context"

	"g.hz.netease.com/horizon/pkg/cluster/code"
)

type Controller interface {
	ListBranch(ctx context.Context, gitURL string) ([]string, error)
}

func NewController(getter code.GitGetter) Controller {
	return &controller{gitGetter: getter}
}

type controller struct {
	gitGetter code.GitGetter
}

func (c *controller) ListBranch(ctx context.Context, gitURL string) ([]string, error) {
	return c.gitGetter.ListBranch(ctx, gitURL)
}
