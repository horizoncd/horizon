package rpcerror

type ErrorCode string

type RPCError struct {
	httpCode     int
	ErrorCode    ErrorCode `json:"errorCode"`
	ErrorMessage string    `json:"errorMessage"`
}

func (e RPCError) WithErrMsg(errorMsg string) RPCError {
	return RPCError{
		httpCode:     e.httpCode,
		ErrorCode:    e.ErrorCode,
		ErrorMessage: errorMsg,
	}
}

var (
	AccessError = RPCError{
		httpCode:  403,
		ErrorCode: "AccessDeny",
	}
	InternalError = RPCError{
		httpCode:  500,
		ErrorCode: "InternalError",
	}
	ParamError = RPCError{
		httpCode:  400,
		ErrorCode: "InvalidParam",
	}
)
