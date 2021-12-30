package common

import "github.com/gin-gonic/gin"

type HorizonErrorCode string

type HorizonRPCError struct {
	httpCode int
	ErrCode  HorizonErrorCode `json:"errCode"`
	ErrMsg   string           `json:"errMsg"`
}

func (e HorizonRPCError) WithErrMsg(errorMsg string) HorizonRPCError {
	return HorizonRPCError{
		httpCode: e.httpCode,
		ErrCode:  e.ErrCode,
		ErrMsg:   errorMsg,
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
