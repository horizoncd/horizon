package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"g.hz.netease.com/horizon/pkg/config/argocd"
	"g.hz.netease.com/horizon/pkg/config/db"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	"g.hz.netease.com/horizon/pkg/config/oidc"
	"g.hz.netease.com/horizon/pkg/config/server"
	"g.hz.netease.com/horizon/pkg/config/tekton"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerConfig     server.Config     `yaml:"serverConfig"`
	DBConfig         db.Config         `yaml:"dbConfig"`
	OIDCConfig       oidc.Config       `yaml:"oidcConfig"`
	GitlabMapper     gitlab.Mapper     `yaml:"gitlabMapper"`
	GitlabRepoConfig gitlab.RepoConfig `yaml:"gitlabRepoConfig"`
	ArgoCDMapper     argocd.Mapper     `yaml:"argoCDMapper"`
	TektonMapper     tekton.Mapper     `yaml:"tektonMapper"`
}

func loadConfig(configFilePath string) (*Config, error) {
	var config Config
	data, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	newArgoCDMapper := argocd.Mapper{}
	for key, v := range config.ArgoCDMapper {
		ks := strings.Split(key, ",")
		for i := 0; i < len(ks); i++ {
			newArgoCDMapper[ks[i]] = v
		}
	}
	config.ArgoCDMapper = newArgoCDMapper

	newTektonMapper := tekton.Mapper{}
	for key, v := range config.TektonMapper {
		ks := strings.Split(key, ",")
		for i := 0; i < len(ks); i++ {
			newTektonMapper[ks[i]] = v
		}
	}
	config.TektonMapper = newTektonMapper

	fmt.Printf("%v", config)
	return &config, nil
}
