package cmd

import "g.hz.netease.com/horizon/pkg/config/db"

type Config struct {
	DBConfig db.Config `yaml:"dbconfig"`
}
