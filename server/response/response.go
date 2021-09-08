package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	ErrorCode    string      `json:"errorCode,omitempty"`
	ErrorMessage string      `json:"errorMessage,omitempty"`
	Data         interface{} `json:"data,omitempty"`
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

func Abort(c *gin.Context, httpCode int, errorCode, errorMessage string) {
	c.JSON(httpCode, &Response{
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
	})
	c.Abort()
}

func AbortWithRequestError(c *gin.Context, errorCode, errorMessage string) {
	Abort(c, http.StatusBadRequest, errorCode, errorMessage)
}

func AbortWithInternalError(c *gin.Context, errorCode, message string) {
	Abort(c, http.StatusInternalServerError, errorCode, message)
}
