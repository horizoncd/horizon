package environment

import (
	"g.hz.netease.com/horizon/core/controller/environmentregion"
)

const (
	// param
	_environmentParam = "environment"
)

type API struct {
	envRegionCtl environmentregion.Controller
}

func NewAPI() *API {
	return &API{
		envRegionCtl: environmentregion.Ctl,
	}
}
