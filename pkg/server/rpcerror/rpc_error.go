package rpcerror

import "net/http"

type ErrorCode string

type RPCError struct {
	HttpCode     int       `json:"-"`
	ErrorCode    ErrorCode `json:"errorCode"`
	ErrorMessage string    `json:"errorMessage"`
}

func (e RPCError) WithErrMsg(errorMsg string) RPCError {
	return RPCError{
		HttpCode:     e.HttpCode,
		ErrorCode:    e.ErrorCode,
		ErrorMessage: errorMsg,
	}
}

var (
	ForbiddenError = RPCError{
		HttpCode:  http.StatusForbidden,
		ErrorCode: "AccessDeny",
	}
	InternalError = RPCError{
		HttpCode:  http.StatusInternalServerError,
		ErrorCode: "InternalError",
	}
	ParamError = RPCError{
		HttpCode:  http.StatusBadRequest,
		ErrorCode: "InvalidParam",
	}
	NotFoundError = RPCError{
		HttpCode:  http.StatusNotFound,
		ErrorCode: "NotFound",
	}
)
