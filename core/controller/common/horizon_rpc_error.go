package common

type HorizonErrorCode string

type HorizonRPCError struct {
	httpCode int
	ErrCode  HorizonErrorCode
	ErrMsg   string
}

func (e *HorizonRPCError) WithErrMsg(errorMsg string) {
	e.ErrMsg = errorMsg
}

var (
	AccessError   = HorizonRPCError{403, "AccessDeny", ""}
	InternalError = HorizonRPCError{500, "InternalError", ""}
)
