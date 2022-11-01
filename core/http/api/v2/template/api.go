package template

import (
	"context"
	"fmt"
	"strconv"

	templatectl "g.hz.netease.com/horizon/core/controller/template"
	herrors "g.hz.netease.com/horizon/core/errors"
	tplctx "g.hz.netease.com/horizon/pkg/context"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
	"github.com/gin-gonic/gin"
)

const (
	_groupParam = "groupID"

	_listRecursively = "recursive"
	_withFullPath    = "fullpath"
)

type API struct {
	templateCtl templatectl.Controller
}

func NewAPI(ctl templatectl.Controller) *API {
	return &API{
		templateCtl: ctl,
	}
}

func (a *API) ListTemplatesByGroupID(c *gin.Context) {
	op := "template: list templates by group id"

	g := c.Param(_groupParam)

	withFullPathStr := c.Query(_withFullPath)
	withFullPath, err := strconv.ParseBool(withFullPathStr)
	var ctx context.Context = c
	if err == nil {
		ctx = context.WithValue(ctx, tplctx.TemplateWithFullPath, withFullPath)
	}

	listRecursivelyStr := c.Query(_listRecursively)
	listRecursively, err := strconv.ParseBool(listRecursivelyStr)
	if err == nil {
		ctx = context.WithValue(ctx, tplctx.TemplateListRecursively, listRecursively)
	}

	var groupID uint64
	if groupID, err = strconv.ParseUint(g, 10, 64); err != nil {
		log.WithFiled(c, "op", op).Info("clusterID not found or invalid")
		response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg("clusterID not found or invalid"))
		return
	}

	var templates templatectl.Templates
	if templates, err = a.templateCtl.ListTemplateByGroupID(ctx, uint(groupID), true); err != nil {
		if perror.Cause(err) == herrors.ErrNoPrivilege {
			log.WithFiled(c, "op", op).Info("non-admin user try to access root group")
			response.AbortWithRPCError(c, rpcerror.ForbiddenError.WithErrMsg(fmt.Sprintf("no privilege: %s", err.Error())))
			return
		}
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			log.WithFiled(c, "op", op).Infof("group with ID %d not found", groupID)
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(fmt.Sprintf("not found: %s", err)))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(fmt.Sprintf("%s", err)))
		return
	}
	response.SuccessWithData(c, templates)
}
