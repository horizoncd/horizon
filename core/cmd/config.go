package cmd

import (
	"io/ioutil"
	"strings"

	"g.hz.netease.com/horizon/pkg/config/argocd"
	"g.hz.netease.com/horizon/pkg/config/authenticate"
	"g.hz.netease.com/horizon/pkg/config/cmdb"
	"g.hz.netease.com/horizon/pkg/config/db"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	"g.hz.netease.com/horizon/pkg/config/grafana"
	"g.hz.netease.com/horizon/pkg/config/helmrepo"
	"g.hz.netease.com/horizon/pkg/config/oidc"
	"g.hz.netease.com/horizon/pkg/config/server"
	"g.hz.netease.com/horizon/pkg/config/tekton"

	"gopkg.in/yaml.v2"
)

type Config struct {
	ServerConfig           server.Config           `yaml:"serverConfig"`
	CloudEventServerConfig server.Config           `yaml:"cloudEventServerConfig"`
	DBConfig               db.Config               `yaml:"dbConfig"`
	OIDCConfig             oidc.Config             `yaml:"oidcConfig"`
	GitlabMapper           gitlab.Mapper           `yaml:"gitlabMapper"`
	GitlabRepoConfig       gitlab.RepoConfig       `yaml:"gitlabRepoConfig"`
	ArgoCDMapper           argocd.Mapper           `yaml:"argoCDMapper"`
	TektonMapper           tekton.Mapper           `yaml:"tektonMapper"`
	HelmRepoMapper         helmrepo.Mapper         `yaml:"helmRepoMapper"`
	AccessSecretKeys       authenticate.KeysConfig `yaml:"accessSecretKeys"`
	CmdbConfig             cmdb.Config             `yaml:"cmdbConfig"`
	GrafanaMapper          grafana.Mapper          `yaml:"grafanaMapper"`
	GrafanaSLO             grafana.SLO             `yaml:"grafanaSLO"`
	OauthHTMLLocation      string                  `yaml:"oauthHTMLLocation"`
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

	newHelmRepoMapper := helmrepo.Mapper{}
	for key, v := range config.HelmRepoMapper {
		ks := strings.Split(key, ",")
		for i := 0; i < len(ks); i++ {
			newHelmRepoMapper[ks[i]] = v
		}
	}
	config.HelmRepoMapper = newHelmRepoMapper

	newGrafanaMapper := grafana.Mapper{}
	for key, v := range config.GrafanaMapper {
		ks := strings.Split(key, ",")
		for i := 0; i < len(ks); i++ {
			newGrafanaMapper[ks[i]] = v
		}
	}
	config.GrafanaMapper = newGrafanaMapper

	return &config, nil
}
