package common

import "github.com/gin-gonic/gin"

type HorizonErrorCode string

type HorizonRPCError struct {
	httpCode  int
	ErrorCode HorizonErrorCode `json:"errorCode"`
	ErrorMsg  string           `json:"errorMessage"`
}

func (e HorizonRPCError) WithErrMsg(errorMsg string) HorizonRPCError {
	return HorizonRPCError{
		httpCode:  e.httpCode,
		ErrorCode: e.ErrorCode,
		ErrorMsg:  errorMsg,
	}
}

func Response(c *gin.Context, err HorizonRPCError) {
	c.JSON(err.httpCode, err)
}

var (
	AccessError   = HorizonRPCError{403, "AccessDeny", ""}
	InternalError = HorizonRPCError{500, "InternalError", ""}
	ParamError    = HorizonRPCError{400, "InvalidParam", ""}
)
