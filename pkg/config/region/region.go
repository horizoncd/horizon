package region

import (
	"io"
	"io/ioutil"
	"strings"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DefaultRegions DefaultRegions `yaml:"defaultRegions"`
}

// DefaultRegions key is environment, value is default region of this environment
type DefaultRegions map[string]string

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

	return config
}
