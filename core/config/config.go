package config

import (
	"fmt"
	"io/ioutil"
	"strings"

	"g.hz.netease.com/horizon/pkg/config/argocd"
	"g.hz.netease.com/horizon/pkg/config/authenticate"
	"g.hz.netease.com/horizon/pkg/config/autofree"
	"g.hz.netease.com/horizon/pkg/config/cmdb"
	"g.hz.netease.com/horizon/pkg/config/db"
	"g.hz.netease.com/horizon/pkg/config/eventhandler"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	"g.hz.netease.com/horizon/pkg/config/grafana"
	"g.hz.netease.com/horizon/pkg/config/oauth"
	"g.hz.netease.com/horizon/pkg/config/pprof"
	"g.hz.netease.com/horizon/pkg/config/redis"
	"g.hz.netease.com/horizon/pkg/config/server"
	"g.hz.netease.com/horizon/pkg/config/session"
	"g.hz.netease.com/horizon/pkg/config/tekton"
	"g.hz.netease.com/horizon/pkg/config/templaterepo"
	"g.hz.netease.com/horizon/pkg/config/webhook"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerConfig           server.Config           `yaml:"serverConfig"`
	CloudEventServerConfig server.Config           `yaml:"cloudEventServerConfig"`
	PProf                  pprof.Config            `yaml:"pprofConfig"`
	DBConfig               db.Config               `yaml:"dbConfig"`
	SessionConfig          session.Config          `yaml:"sessionConfig"`
	GitlabMapper           gitlab.Mapper           `yaml:"gitlabMapper"`
	GitopsRepoConfig       gitlab.GitopsRepoConfig `yaml:"gitopsRepoConfig"`
	ArgoCDMapper           argocd.Mapper           `yaml:"argoCDMapper"`
	RedisConfig            redis.Redis             `yaml:"redisConfig"`
	TektonMapper           tekton.Mapper           `yaml:"tektonMapper"`
	TemplateRepo           templaterepo.Repo       `yaml:"templateRepo"`
	AccessSecretKeys       authenticate.KeysConfig `yaml:"accessSecretKeys"`
	CmdbConfig             cmdb.Config             `yaml:"cmdbConfig"`
	GrafanaConfig          grafana.Config          `yaml:"grafanaConfig"`
	Oauth                  oauth.Server            `yaml:"oauth"`
	AutoFreeConfig         autofree.Config         `yaml:"autoFree"`
	KubeConfig             string                  `yaml:"kubeconfig"`
	WebhookConfig          webhook.Config          `yaml:"webhook"`
	EventHandlerConfig     eventhandler.Config     `yaml:"eventHandler"`
}

func LoadConfig(configFilePath string) (*Config, error) {
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

	// use http schema by default
	if config.GitopsRepoConfig.URLSchema == "" {
		config.GitopsRepoConfig.URLSchema = gitlab.HTTPURLSchema
	}

	// validte gitops repo schema value
	if config.GitopsRepoConfig.URLSchema != gitlab.HTTPURLSchema &&
		config.GitopsRepoConfig.URLSchema != gitlab.SSHURLSchema {
		return nil, fmt.Errorf("gitops repo urlSchema not valid: %s, valid values: [http, ssh]",
			config.GitopsRepoConfig.URLSchema)
	}

	return &config, nil
}
