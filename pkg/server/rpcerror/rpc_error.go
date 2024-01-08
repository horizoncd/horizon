// Copyright © 2023 Horizoncd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rpcerror

import (
	"fmt"
	"net/http"
)

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

func (e RPCError) WithErrMsgf(format string, params ...interface{}) RPCError {
	return RPCError{
		HTTPCode:     e.HTTPCode,
		ErrorCode:    e.ErrorCode,
		ErrorMessage: fmt.Sprintf(format, params...),
	}
}

func (e RPCError) Error() string {
	return fmt.Sprintf("HTTPCode: %d, ErrorCode: %s, ErrorMessage: %s", e.HTTPCode, e.ErrorCode, e.ErrorMessage)
}

var (
	ForbiddenError = RPCError{
		HTTPCode:  http.StatusForbidden,
		ErrorCode: "AccessDeny",
	}
	Unauthorized = RPCError{
		HTTPCode:  http.StatusUnauthorized,
		ErrorCode: "Unauthorized",
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
