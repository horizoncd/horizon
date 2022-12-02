package handler

import (
	"errors"
	"fmt"
	"reflect"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/application"
	"g.hz.netease.com/horizon/core/controller/cluster"
	applicationmodel "g.hz.netease.com/horizon/pkg/application/models"
	"g.hz.netease.com/horizon/pkg/cmdb"
	"g.hz.netease.com/horizon/pkg/hook/hook"
)

func NewCMDBEventHandler(controller cmdb.Controller) EventHandler {
	return &CMDBEventHandler{ctl: controller}
}

type CMDBEventHandler struct {
	ctl cmdb.Controller
}

func (h *CMDBEventHandler) Process(event *hook.EventCtx) error {
	switch event.EventType {
	case hook.CreateApplication:
		return h.ProcessCreateApplication(event)
	case hook.DeleteApplication:
		return h.ProcessDeleteApplication(event)
	case hook.CreateCluster:
		return h.ProcessCreateCluster(event)
	case hook.DeleteCluster:
		return h.ProcessDeleteCluster(event)
	default:
		return fmt.Errorf("unsupported eventType %s", event.EventType)
	}
}

func (h *CMDBEventHandler) ProcessCreateApplication(event *hook.EventCtx) error {
	currentUser, err := common.UserFromContext(event.Ctx)
	if err != nil {
		return errors.New("can not get user from context")
	}
	var (
		priority string
		name     string
	)

	e := event.Event
	switch info := e.(type) {
	// nolint
	case *application.GetApplicationResponse:
		priority = info.Priority
		name = info.Name
	// nolint
	case *application.CreateApplicationResponseV2:
		priority = info.Priority
		name = info.Name
	default:
		return fmt.Errorf("type error, need GetApplicationResponse | "+
			"CreateApplicationResponseV2, get %s", reflect.TypeOf(e).Name())
	}

	accounts := make([]cmdb.Account, 0)
	accounts = append(accounts, cmdb.Account{
		Account:     currentUser.GetName(),
		AccountType: cmdb.User,
	})

	var cmdbPriority cmdb.PriorityType
	switch applicationmodel.Priority(priority) {
	case applicationmodel.P0:
		cmdbPriority = cmdb.P0
	case applicationmodel.P1:
		cmdbPriority = cmdb.P1
	case applicationmodel.P2:
		cmdbPriority = cmdb.P2
	default:
		cmdbPriority = cmdb.P2
	}
	req := cmdb.CreateApplicationRequest{
		Name:     name,
		Priority: cmdbPriority,
		Admin:    accounts,
	}
	return h.ctl.CreateApplication(event.Ctx, req)
}

func (h *CMDBEventHandler) ProcessDeleteApplication(event *hook.EventCtx) error {
	e := event.Event
	switch info := e.(type) {
	case string:
		return h.ctl.DeleteApplication(event.Ctx, info)
	default:
		return fmt.Errorf("type error,need string, get %s", reflect.TypeOf(e).Name())
	}
}

func (h *CMDBEventHandler) ProcessCreateCluster(event *hook.EventCtx) error {
	currentUser, err := common.UserFromContext(event.Ctx)
	if err != nil {
		return errors.New("can not get user from context")
	}
	var (
		environment     string
		name            string
		applicationName string
	)

	e := event.Event
	switch info := e.(type) {
	// nolint
	case *cluster.GetClusterResponse:
		environment = info.Scope.Environment
		name = info.Name
		applicationName = info.Application.Name
	// nolint
	case *cluster.CreateClusterResponseV2:
		environment = info.Scope.Environment
		name = info.Name
		applicationName = info.Application.Name
	default:
		return fmt.Errorf("type error, need GetClusterResponse | CreateClusterResponseV2, get %s", reflect.TypeOf(e).Name())
	}

	accounts := make([]cmdb.Account, 0)
	accounts = append(accounts, cmdb.Account{
		Account:     currentUser.GetName(),
		AccountType: cmdb.User,
	})

	env, err := cmdb.ToCmdbEnv(environment)
	if err != nil {
		return err
	}
	req := cmdb.CreateClusterRequest{
		Name:                name,
		ApplicationName:     applicationName,
		Env:                 env,
		ClusterServerStatus: cmdb.StatusReady,
		AutoAddDocker:       cmdb.AutoAddContainer,
		Admin:               accounts,
	}
	return h.ctl.CreateCluster(event.Ctx, req)
}

func (h *CMDBEventHandler) ProcessDeleteCluster(event *hook.EventCtx) error {
	e := event.Event
	switch info := e.(type) {
	case string:
		return h.ctl.DeleteCluster(event.Ctx, info)
	default:
		return fmt.Errorf("type error,need string, get %s", reflect.TypeOf(e).Name())
	}
}
