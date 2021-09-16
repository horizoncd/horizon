package errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

type Op string

const (
	ENil        = 0
	EBadRequest = http.StatusBadRequest
	ENotFound   = http.StatusNotFound
	EInternal   = http.StatusInternalServerError

	EDuplicate = 600
)

type Error struct {
	// code for machine-readable error code.
	code int
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
		case int:
			e.code = arg
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
			_, _ = fmt.Fprintf(&builder, "%v: ", ne.op)
		}
	}
	for err := error(e); err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok {
			if e.code != ENil || e.msg != "" {
				code := e.code
				if e.code == ENil {
					code = EInternal
				}
				_, _ = fmt.Fprintf(&builder, "<%v %v> ", code, e.msg)
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

// ErrorCode returns the code of the root error, if available. Otherwise returns EInternal.
func ErrorCode(err error) int {
	for ; err != nil; err = errors.Unwrap(err) {
		if e, ok := err.(*Error); ok && e.code != ENil {
			return e.code
		}
	}
	return EInternal
}

func ErrorMsg(err error) string {
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
