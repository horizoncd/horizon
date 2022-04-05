package models

import (
	"g.hz.netease.com/horizon/pkg/server/global"
)

type Harbor struct {
	global.Model

	Server          string
	Token           string
	PreheatPolicyID int
}
