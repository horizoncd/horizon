// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package registry

import (
	"context"

	herrors "github.com/horizoncd/horizon/core/errors"
	perror "github.com/horizoncd/horizon/pkg/errors"
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
