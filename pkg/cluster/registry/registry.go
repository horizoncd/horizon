package registry

import (
	"context"

	herrors "g.hz.netease.com/horizon/core/errors"
	perror "g.hz.netease.com/horizon/pkg/errors"
)

type Constructor func(config *Config) (Registry, error)

var factory = make(map[string]Constructor)

func GetKinds() []string {
	kinds := make([]string, 0, len(factory))
	for kind := range factory {
		kinds = append(kinds, kind)
	}
	return kinds
}

func Register(kind string, constructor Constructor) {
	factory[kind] = constructor
}

// Registry ...
//
// nolint
//
//go:generate mockgen -source=$GOFILE -destination=../../../mock/pkg/cluster/registry/registry_mock.go -package=mock_registry
type Registry interface {
	// DeleteImage delete repository
	DeleteImage(ctx context.Context, appName string, clusterName string) error
}

type Config struct {
	Server             string
	Token              string
	Path               string
	InsecureSkipVerify bool

	Kind string
}

func NewRegistry(config *Config) (Registry, error) {
	for kind, constructor := range factory {
		if kind == config.Kind {
			return constructor(config)
		}
	}
	return nil, perror.Wrapf(herrors.ErrParamInvalid, "kind = %v is not implement", config.Kind)
}
