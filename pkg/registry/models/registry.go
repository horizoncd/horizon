package models

import (
	"github.com/horizoncd/horizon/pkg/server/global"
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
