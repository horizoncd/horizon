package rpcerror

import "net/http"

type ErrorCode string

type RPCError struct {
	HTTPCode     int       `json:"-"`
	ErrorCode    ErrorCode `json:"errorCode"`
	ErrorMessage string    `json:"errorMessage"`
}

func (e RPCError) WithErrMsg(errorMsg string) RPCError {
	return RPCError{
		HTTPCode:     e.HTTPCode,
		ErrorCode:    e.ErrorCode,
		ErrorMessage: errorMsg,
	}
}

var (
	ForbiddenError = RPCError{
		HTTPCode:  http.StatusForbidden,
		ErrorCode: "AccessDeny",
	}
	InternalError = RPCError{
		HTTPCode:  http.StatusInternalServerError,
		ErrorCode: "InternalError",
	}
	ParamError = RPCError{
		HTTPCode:  http.StatusBadRequest,
		ErrorCode: "InvalidParam",
	}
	BadRequestError = RPCError{
		HTTPCode:  http.StatusBadRequest,
		ErrorCode: "Bad Request",
	}
	NotFoundError = RPCError{
		HTTPCode:  http.StatusNotFound,
		ErrorCode: "NotFound",
	}
	ConflictError = RPCError{
		HTTPCode:  http.StatusConflict,
		ErrorCode: "Conflict",
	}
)
