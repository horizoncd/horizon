package common

type HorizonErrorCode string

type HorizonError struct {
	httpCode int
	ErrCode  HorizonErrorCode
	ErrMsg   string
}

func (e *HorizonError) WithErrMsg(errorMsg string) {
	e.ErrMsg = errorMsg
}

var (
	AccessError   = HorizonError{403, "AccessDeny", ""}
	InternalError = HorizonError{500, "InternalError", ""}
)
