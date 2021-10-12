package cmd

import (
	"g.hz.netease.com/horizon/pkg/config/db"
	"g.hz.netease.com/horizon/pkg/config/gitlab"
	"g.hz.netease.com/horizon/pkg/config/oidc"
	"g.hz.netease.com/horizon/pkg/config/server"
)

type Config struct {
	ServerConfig server.Config `yaml:"serverConfig"`
	DBConfig     db.Config     `yaml:"dbConfig"`
	OIDCConfig   oidc.Config   `yaml:"oidcConfig"`
	GitlabConfig gitlab.Config `yaml:"gitlabConfig"`
}
