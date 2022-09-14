package webhook

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"g.hz.netease.com/horizon/core/common"
	"g.hz.netease.com/horizon/core/controller/webhook"
	webhookctl "g.hz.netease.com/horizon/core/controller/webhook"
	herrors "g.hz.netease.com/horizon/core/errors"
	"g.hz.netease.com/horizon/lib/q"
	perror "g.hz.netease.com/horizon/pkg/errors"
	"g.hz.netease.com/horizon/pkg/server/response"
	"g.hz.netease.com/horizon/pkg/server/rpcerror"
	"g.hz.netease.com/horizon/pkg/util/log"
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
	request.ResourceType = resourceType
	request.ResourceID = uint(resourceID)

	resp, err := a.webhookCtl.CreateWebhook(c, &request)
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

	var (
		pageNumber, pageSize int
	)

	pageNumberStr := c.Query(common.PageNumber)
	if pageNumberStr == "" {
		pageNumber = common.DefaultPageNumber
	} else {
		pageNumber, err = strconv.Atoi(pageNumberStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid pageNumber")
			return
		}
	}

	pageSizeStr := c.Query(common.PageSize)
	if pageSizeStr == "" {
		pageSize = common.DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid pageSize")
			return
		}
	}

	items, total, err := a.webhookCtl.ListWebhooks(c, resourceType, uint(resourceID), &q.Query{
		PageSize:   pageSize,
		PageNumber: pageNumber,
	})
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

	var (
		pageNumber int
		pageSize   int
	)

	pageSizeStr := c.Query(common.PageSize)
	if pageSizeStr == "" {
		pageSize = common.DefaultPageSize
	} else {
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil {
			response.AbortWithRequestError(c, common.InvalidRequestParam, "invalid pageSize")
			return
		}
	}

	items, total, err := a.webhookCtl.ListWebhookLogs(c, uint(webhookID), &q.Query{
		PageNumber: pageNumber,
		PageSize:   pageSize,
	})
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

func (a *API) RetryWebhookLog(c *gin.Context) {
	const op = "webhook: retry log"
	idStr := c.Param(_webhookLogIDParam)
	id, err := strconv.ParseUint(idStr, 10, 0)
	if err != nil {
		response.AbortWithRPCError(c, rpcerror.ParamError.
			WithErrMsgf("invalid id: %s", idStr))
		return
	}

	resp, err := a.webhookCtl.RetryWebhookLog(c, uint(id))
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
