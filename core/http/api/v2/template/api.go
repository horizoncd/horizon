package template

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/horizoncd/horizon/core/common"
	templatectl "github.com/horizoncd/horizon/core/controller/template"
	herrors "github.com/horizoncd/horizon/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	tplctx "github.com/horizoncd/horizon/pkg/context"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
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

func (a *API) List(c *gin.Context) {
	withFullPathStr := c.Query(_withFullPath)
	withFullPath, _ := strconv.ParseBool(withFullPathStr)

	keywords := q.KeyWords{}

	userIDStr := c.Query(common.TemplateQueryByUser)
	if userIDStr != "" {
		userID, err := strconv.ParseUint(userIDStr, 10, 0)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsgf(
					"failed to parse userID\n"+
						"userID = %s\nerr = %v", userIDStr, err))
			return
		}
		keywords[common.TemplateQueryByUser] = uint(userID)
	}

	groupIDStr := c.Query(common.TemplateQueryByGroupRecursive)
	if groupIDStr != "" {
		groupID, err := strconv.ParseUint(groupIDStr, 10, 0)
		if err != nil {
			response.AbortWithRPCError(c,
				rpcerror.ParamError.WithErrMsgf(
					"failed to parse groupID\n"+
						"groupID = %s\nerr = %v", groupIDStr, err))
			return
		}
		keywords[common.TemplateQueryByGroupRecursive] = uint(groupID)
	}

	filter := c.Query(common.TemplateQueryName)
	if filter != "" {
		keywords[common.TemplateQueryName] = filter
	}

	query := q.New(keywords).WithPagination(c)

	total, templates, err := a.templateCtl.ListV2(c, query, withFullPath)
	if err != nil {
		if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(
				c, rpcerror.NotFoundError.WithErrMsgf("templates not found: %v", err),
			)
			return
		}
		response.AbortWithRPCError(c,
			rpcerror.InternalError.WithErrMsgf("failed to get templates: %v", err))
	}
	response.SuccessWithData(c, response.DataWithTotal{
		Total: int64(total),
		Items: templates,
	})
}
