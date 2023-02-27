package config

import (
	"io/ioutil"
	"strings"

	"github.com/horizoncd/horizon/pkg/config/argocd"
	"github.com/horizoncd/horizon/pkg/config/authenticate"
	"github.com/horizoncd/horizon/pkg/config/autofree"
	"github.com/horizoncd/horizon/pkg/config/db"
	"github.com/horizoncd/horizon/pkg/config/eventhandler"
	"github.com/horizoncd/horizon/pkg/config/git"
	"github.com/horizoncd/horizon/pkg/config/gitlab"
	"github.com/horizoncd/horizon/pkg/config/grafana"
	"github.com/horizoncd/horizon/pkg/config/oauth"
	"github.com/horizoncd/horizon/pkg/config/pprof"
	"github.com/horizoncd/horizon/pkg/config/redis"
	"github.com/horizoncd/horizon/pkg/config/server"
	"github.com/horizoncd/horizon/pkg/config/session"
	"github.com/horizoncd/horizon/pkg/config/tekton"
	"github.com/horizoncd/horizon/pkg/config/template"
	"github.com/horizoncd/horizon/pkg/config/templaterepo"
	"github.com/horizoncd/horizon/pkg/config/token"
	"github.com/horizoncd/horizon/pkg/config/webhook"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerConfig           server.Config           `yaml:"serverConfig"`
	CloudEventServerConfig server.Config           `yaml:"cloudEventServerConfig"`
	PProf                  pprof.Config            `yaml:"pprofConfig"`
	DBConfig               db.Config               `yaml:"dbConfig"`
	SessionConfig          session.Config          `yaml:"sessionConfig"`
	GitopsRepoConfig       gitlab.GitopsRepoConfig `yaml:"gitopsRepoConfig"`
	ArgoCDMapper           argocd.Mapper           `yaml:"argoCDMapper"`
	RedisConfig            redis.Redis             `yaml:"redisConfig"`
	TektonMapper           tekton.Mapper           `yaml:"tektonMapper"`
	TemplateRepo           templaterepo.Repo       `yaml:"templateRepo"`
	AccessSecretKeys       authenticate.KeysConfig `yaml:"accessSecretKeys"`
	GrafanaConfig          grafana.Config          `yaml:"grafanaConfig"`
	Oauth                  oauth.Server            `yaml:"oauth"`
	AutoFreeConfig         autofree.Config         `yaml:"autoFree"`
	KubeConfig             string                  `yaml:"kubeconfig"`
	WebhookConfig          webhook.Config          `yaml:"webhook"`
	EventHandlerConfig     eventhandler.Config     `yaml:"eventHandler"`
	CodeGitRepos           []*git.Repo             `yaml:"gitRepos"`
	TokenConfig            token.Config            `yaml:"tokenConfig"`
	TemplateUpgradeMapper  template.UpgradeMapper  `yaml:"templateUpgradeMapper"`
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

	if config.EventHandlerConfig.BatchEventsCount <= 0 {
		config.EventHandlerConfig.BatchEventsCount = 5
	}
	if config.EventHandlerConfig.CursorSaveInterval <= 0 {
		config.EventHandlerConfig.CursorSaveInterval = 10
	}
	if config.EventHandlerConfig.IdleWaitInterval <= 0 {
		config.EventHandlerConfig.IdleWaitInterval = 3
	}
	if config.WebhookConfig.ClientTimeout <= 0 {
		config.WebhookConfig.ClientTimeout = 30
	}
	if config.WebhookConfig.IdleWaitInterval <= 0 {
		config.WebhookConfig.IdleWaitInterval = 2
	}
	if config.WebhookConfig.WorkerReconcileInterval <= 0 {
		config.WebhookConfig.WorkerReconcileInterval = 5
	}
	if config.WebhookConfig.ResponseBodyTruncateSize <= 0 {
		config.WebhookConfig.ResponseBodyTruncateSize = 16384
	}

	return &config, nil
}
