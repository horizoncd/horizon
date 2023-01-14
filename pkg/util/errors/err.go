// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errors

import (
	"errors"
	"fmt"
	"strings"
)

type (
	Op        string
	ErrorCode string
)

const (
	ErrCodeInternalError = ErrorCode("InternalError")

	StatusInternalError = 500
)

type Error struct {
	// status for status code
	status int
	// code for machine-readable error code.
	code ErrorCode
	// msg for human-readable message.
	msg string
	// Op and err for logical operation and nested error.
	op  Op
	err error
}

func E(op Op, args ...interface{}) error {
	e := &Error{}
	e.op = op
	for _, arg := range args {
		switch arg := arg.(type) {
		case error:
			e.err = arg
		case string:
			e.msg = arg
		case ErrorCode:
			e.code = arg
		case int:
			e.status = arg
		default:
			panic("bad call to E")
		}
	}
	return e
}

func (e *Error) Unwrap() error {
	return e.err
}

// Error returns the string representation of the error message.
func (e *Error) Error() string {
	var builder strings.Builder
	for err := error(e); err != nil; err = errors.Unwrap(err) {
		if ne, ok := err.(*Error); ok {
			_, _ = fmt.Fprintf(&builder, "%v - ", ne.op)
		}
	}
	for err := error(e); err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok {
			if e.status != 0 || e.code != "" || e.msg != "" {
				code := e.code
				if e.code == "" {
					code = ErrCodeInternalError
				}
				status := e.status
				if e.status == 0 {
					status = StatusInternalError
				}
				_, _ = fmt.Fprintf(&builder, "<%v %v - %v> ", status, code, e.msg)
			}
		} else {
			str := builder.String()
			if len(str) > 1 && str[len(str)-1] == ' ' {
				_, _ = fmt.Fprintf(&builder, "%v", err)
			} else {
				_, _ = fmt.Fprintf(&builder, " %v", err)
			}
		}
	}
	return builder.String()
}

// Status return the status of the root error, if available. Otherwise, returns the StatusInternalError.
func Status(err error) int {
	for ; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok && e.status != 0 {
			return e.status
		}
	}
	return StatusInternalError
}

// Code returns the code of the root error, if available. Otherwise, returns ErrCodeInternalError.
func Code(err error) string {
	for ; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok && e.code != "" {
			return string(e.code)
		}
	}
	return string(ErrCodeInternalError)
}

// Message returns the message of the root error, if available.
func Message(err error) string {
	for ; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok {
			if e.msg != "" {
				return e.msg
			}
		} else {
			return err.Error()
		}
	}
	return "An internal error has occurred. Please contact technical support."
}
