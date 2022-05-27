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
	e := event.Event
	switch info := e.(type) {
	case *application.GetApplicationResponse:
		currentUser, err := common.FromContext(event.Ctx)
		if err != nil {
			return errors.New("can not get user from context")
		}
		accounts := make([]cmdb.Account, 0)
		accounts = append(accounts, cmdb.Account{
			Account:     currentUser.GetName(),
			AccountType: cmdb.User,
		})

		var priority cmdb.PriorityType
		switch applicationmodel.Priority(info.Priority) {
		case applicationmodel.P0:
			priority = cmdb.P0
		case applicationmodel.P1:
			priority = cmdb.P1
		case applicationmodel.P2:
			priority = cmdb.P2
		default:
			priority = cmdb.P2
		}
		req := cmdb.CreateApplicationRequest{
			Name:     info.Name,
			Priority: priority,
			Admin:    accounts,
		}
		return h.ctl.CreateApplication(event.Ctx, req)
	default:
		return fmt.Errorf("type error,need GetApplicationResponse, get %s", reflect.TypeOf(e).Name())
	}
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
	e := event.Event
	switch info := e.(type) {
	case *cluster.GetClusterResponse:
		currentUser, err := common.FromContext(event.Ctx)
		if err != nil {
			return errors.New("can not get user from context")
		}
		accounts := make([]cmdb.Account, 0)
		accounts = append(accounts, cmdb.Account{
			Account:     currentUser.GetName(),
			AccountType: cmdb.User,
		})

		env, err := cmdb.ToCmdbEnv(info.Scope.Environment)
		if err != nil {
			return err
		}
		req := cmdb.CreateClusterRequest{
			Name:                info.Name,
			ApplicationName:     info.Application.Name,
			Env:                 env,
			ClusterServerStatus: cmdb.StatusReady,
			AutoAddDocker:       cmdb.AutoAddContainer,
			Admin:               accounts,
		}
		return h.ctl.CreateCluster(event.Ctx, req)
	default:
		return fmt.Errorf("type error,need string, get %s", reflect.TypeOf(e).Name())
	}
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
