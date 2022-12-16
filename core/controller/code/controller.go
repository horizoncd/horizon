package code

import (
	"context"

	"github.com/horizoncd/horizon/pkg/cluster/code"
)

type Controller interface {
	ListBranch(ctx context.Context, gitURL string, params *code.SearchParams) ([]string, error)
	ListTag(ctx context.Context, gitURL string, params *code.SearchParams) ([]string, error)
}

func NewController(getter code.GitGetter) Controller {
	return &controller{gitGetter: getter}
}

type controller struct {
	gitGetter code.GitGetter
}

func (c *controller) ListBranch(ctx context.Context, gitURL string, params *code.SearchParams) ([]string, error) {
	return c.gitGetter.ListBranch(ctx, gitURL, params)
}

func (c *controller) ListTag(ctx context.Context, gitURL string, params *code.SearchParams) ([]string, error) {
	return c.gitGetter.ListTag(ctx, gitURL, params)
}
