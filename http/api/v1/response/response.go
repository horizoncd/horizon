package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	RequestID string      `json:"requestID,omitempty"`
}

func NewResponse(code int, message string) *Response {
	return &Response{
		Code:    code,
		Message: message,
	}
}

func NewResponseWithData(code int, message string, data interface{}) *Response {
	return &Response{
		Code:    code,
		Message: message,
		Data:    data,
	}
}

func Success(c *gin.Context) {
	c.JSON(http.StatusOK, NewResponse(http.StatusOK, http.StatusText(http.StatusOK)))
}

func SuccessWithData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, NewResponseWithData(http.StatusOK, http.StatusText(http.StatusOK), data))
}

func abort(c *gin.Context, code int, message string) {
	c.JSON(code, &Response{
		Code:    code,
		Message: message,
	})
	c.Abort()
}

func AbortWithRequestError(c *gin.Context, message string) {
	abort(c, http.StatusBadRequest, message)
}

func AbortWithInternalError(c *gin.Context) {
	abort(c, http.StatusInternalServerError, "服务器内部错误")
}
