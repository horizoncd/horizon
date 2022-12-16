package response

import (
	"net/http"

	"github.com/horizoncd/horizon/core/common"
	"github.com/horizoncd/horizon/core/middleware/requestid"
	"github.com/horizoncd/horizon/pkg/server/rpcerror"
	"github.com/horizoncd/horizon/pkg/util/errors"
	"github.com/horizoncd/horizon/pkg/util/log"

	"github.com/gin-gonic/gin"
)

type DataWithTotal struct {
	Total int64       `json:"total"`
	Items interface{} `json:"items"`
}

type Response struct {
	ErrorCode    string      `json:"errorCode,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
	Data         interface{} `json:"data,omitempty"`
	RequestID    string      `json:"requestID,omitempty"`
}

func NewResponse() *Response {
	return &Response{}
}

func NewResponseWithData(data interface{}) *Response {
	return &Response{
		Data: data,
	}
}

func Success(c *gin.Context) {
	c.JSON(http.StatusOK, NewResponse())
}

func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, NewResponseWithData(data))
}

func SuccessWithDataV2(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

func Abort(c *gin.Context, httpCode int, errorCode, errorMessage string) {
	rid, err := requestid.FromContext(c)
	if err != nil {
		log.Errorf(c, "error to get requestID from context, err: %v", err)
	}

	c.JSON(httpCode, &Response{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		RequestID:    rid,
	})
	c.Abort()
}

func AbortWithForbiddenError(c *gin.Context, code, errMessage string) {
	Abort(c, http.StatusForbidden, code, errMessage)
}

func AbortWithUnauthorized(c *gin.Context, code, errMessage string) {
	Abort(c, http.StatusUnauthorized, code, errMessage)
}

func AbortWithRequestError(c *gin.Context, errorCode, errorMessage string) {
	Abort(c, http.StatusBadRequest, errorCode, errorMessage)
}

func AbortWithInternalError(c *gin.Context, message string) {
	Abort(c, http.StatusInternalServerError, common.InternalError, message)
}

func AbortWithNotExistError(c *gin.Context, message string) {
	Abort(c, http.StatusNotFound, common.NotFound, message)
}

func AbortWithRPCError(c *gin.Context, rpcError rpcerror.RPCError) {
	Abort(c, rpcError.HTTPCode, string(rpcError.ErrorCode), rpcError.ErrorMessage)
}

// AbortWithError TODO: remove this function after all error changed to rpcerror.RPCError
func AbortWithError(c *gin.Context, err error) {
	Abort(c, errors.Status(err), errors.Code(err), err.Error())
}
