package webhook

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/horizoncd/horizon/pkg/core/common"
	"github.com/horizoncd/horizon/pkg/core/controller/webhook"
	webhookctl "github.com/horizoncd/horizon/pkg/core/controller/webhook"
	herrors "github.com/horizoncd/horizon/pkg/core/errors"
	"github.com/horizoncd/horizon/lib/q"
	perror "github.com/horizoncd/horizon/pkg/errors"
	"github.com/horizoncd/horizon/pkg/server/response"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/log"
)

type API struct {
	webhookCtl webhookctl.Controller
}

func NewAPI(ctl webhookctl.Controller) *API {
	return &API{
		webhookCtl: ctl,
	}
}

func (a *API) CreateWebhook(c *gin.Context) {
	const op = "webhook: create"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid resource id: %s", resourceIDStr))
		return
	}

	var request webhook.CreateWebhookRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid request body, err: %s", err.Error()))
		return
	}

	resp, err := a.webhookCtl.CreateWebhook(c, resourceType, uint(resourceID), &request)
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) ListWebhooks(c *gin.Context) {
	const op = "webhook: list"
	resourceType := c.Param(_resourceTypeParam)
	resourceIDStr := c.Param(_resourceIDParam)
	resourceID, err := strconv.ParseUint(resourceIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid resource id: %s", resourceIDStr))
		return
	}

	query := q.New(nil).WithPagination(c)
	items, total, err := a.webhookCtl.ListWebhooks(c, resourceType, uint(resourceID), query)
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, response.DataWithTotal{
		Items: items,
		Total: total,
	})
}

func (a *API) UpdateWebhook(c *gin.Context) {
	const op = "webhook: update"
	idStr := c.Param(_webhookIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid id: %s", idStr))
		return
	}

	var request webhook.UpdateWebhookRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid request body, err: %s", err.Error()))
		return
	}

	resp, err := a.webhookCtl.UpdateWebhook(c, uint(id), &request)
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) GetWebhook(c *gin.Context) {
	const op = "webhook: get"
	idStr := c.Param(_webhookIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid id: %s", idStr))
		return
	}

	resp, err := a.webhookCtl.GetWebhook(c, uint(id))
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) DeleteWebhook(c *gin.Context) {
	const op = "webhook: delete"
	idStr := c.Param(_webhookIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid id: %s", idStr))
		return
	}

	err = a.webhookCtl.DeleteWebhook(c, uint(id))
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.Success(c)
}

func (a *API) ListWebhookLogs(c *gin.Context) {
	const op = "webhook: list logs"
	webhookIDStr := c.Param(_webhookIDParam)
	webhookID, err := strconv.ParseUint(webhookIDStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid webhook id: %s", webhookIDStr))
		return
	}

	keywords := q.KeyWords{}
	if eventType := c.Query(common.EventType); eventType != "" {
		keywords[common.EventType] = c.Query(common.EventType)
	}
	if filter := c.Query(common.Filter); filter != "" {
		keywords[common.Filter] = c.Query(common.Filter)
	}

	query := q.New(keywords).WithPagination(c)
	items, total, err := a.webhookCtl.ListWebhookLogs(c, uint(webhookID), query)
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, response.DataWithTotal{
		Items: items,
		Total: total,
	})
}

func (a *API) GetWebhookLog(c *gin.Context) {
	const op = "webhook: get log"
	idStr := c.Param(_webhookLogIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid id: %s", idStr))
		return
	}

	resp, err := a.webhookCtl.GetWebhookLog(c, uint(id))
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}

func (a *API) ResendWebhook(c *gin.Context) {
	const op = "webhook: resend"
	idStr := c.Param(_webhookLogIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid id: %s", idStr))
		return
	}

	resp, err := a.webhookCtl.ResendWebhook(c, uint(id))
	if err != nil {
		if perror.Cause(err) == herrors.ErrParamInvalid {
			response.AbortWithRPCError(c, rpcerror.ParamError.WithErrMsg(err.Error()))
			return
		} else if _, ok := perror.Cause(err).(*herrors.HorizonErrNotFound); ok {
			response.AbortWithRPCError(c, rpcerror.NotFoundError.WithErrMsg(err.Error()))
			return
		}
		log.WithFiled(c, "op", op).Errorf("%+v", err)
		response.AbortWithRPCError(c, rpcerror.InternalError.WithErrMsg(err.Error()))
		return
	}
	response.SuccessWithData(c, resp)
}
