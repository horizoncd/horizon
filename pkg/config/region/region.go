package region

import (
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DefaultRegions     DefaultRegions     `yaml:"defaultRegions"`
	ApplicationRegions ApplicationRegions `yaml:"applicationRegions"`
}

// DefaultRegions key is environment, value is default region of this environment
type DefaultRegions map[string]string

// ApplicationRegions key is environment, value is a map which key is application and value is its region
type ApplicationRegions map[string]map[string]string

func LoadRegionConfig(reader io.Reader) (*Config, error) {
	var config Config
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return constructConfig(&config), nil
}

func constructConfig(config *Config) *Config {
	newDefaultRegions := DefaultRegions{}
	for key, v := range config.DefaultRegions {
		ks := strings.Split(key, ",")
		for i := 0; i < len(ks); i++ {
			newDefaultRegions[ks[i]] = v
		}
	}
	config.DefaultRegions = newDefaultRegions

	newApplicationRegions := ApplicationRegions{}
	for key, value := range config.ApplicationRegions {
		envs := strings.Split(key, ",")
		for k, v := range value {
			applications := strings.Split(k, ",")
			for i := 0; i < len(envs); i++ {
				env := strings.TrimSpace(envs[i])
				for j := 0; j < len(applications); j++ {
					application := strings.TrimSpace(applications[j])
					if newApplicationRegions[envs[i]] == nil {
						newApplicationRegions[env] = map[string]string{
							application: v,
						}
					} else {
						newApplicationRegions[env][application] = v
					}
				}
			}
		}
	}
	config.ApplicationRegions = newApplicationRegions

	return config
}
