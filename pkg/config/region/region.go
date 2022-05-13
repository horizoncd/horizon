package region

import (
	"context"

	environmentmanager "g.hz.netease.com/horizon/pkg/environment/manager"
)

type Config struct {
	DefaultRegions DefaultRegions `yaml:"defaultRegions"`
}

// DefaultRegions key is environment, value is default region of this environment
type DefaultRegions map[string]string

func FormatEnvironmentDefaultRegions(ctx context.Context) (*Config, error) {
	envs, err := environmentmanager.Mgr.ListAllEnvironment(ctx)
	if err != nil {
		return nil, err
	}

	newDefaultRegions := DefaultRegions{}
	for _, v := range envs {
		if v.DefaultRegion != "" {
			newDefaultRegions[v.Name] = v.DefaultRegion
		}
	}

	return &Config{
		DefaultRegions: newDefaultRegions,
	}, nil
}
