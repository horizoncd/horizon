package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Registry struct {
	global.Model

	Name   string
	Server string
	Path   string
	Token  string
	// for delete
	InsecureSkipTLSVerify bool `gorm:"column:insecure_skip_tls_verify"`
	Kind                  string
}
