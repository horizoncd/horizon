package registry

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
)

type Constructor func(config *Config) (Registry, error)

var Factory = make(map[string]Constructor)

func Register(kind string, constructor Constructor) {
	Factory[kind] = constructor
}

// Registry ...
//
// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/registry/registry_mock.go -package=mock_registry
type Registry interface {
	// DeleteRepository delete repository
	DeleteRepository(ctx context.Context, repository string) error
}

type Config struct {
	Server             string
	Token              string
	Path               string
	InsecureSkipVerify bool

	Kind string
}

func NewRegistry(config *Config) (Registry, error) {
	for kind, constructor := range Factory {
		if kind == config.Kind {
			return constructor(config)
		}
	}
	return nil, perror.Wrapf(herrors.ErrParamInvalid, "kind = %v is not implement", config.Kind)
}
