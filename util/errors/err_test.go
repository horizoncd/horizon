package errors

import (
	"errors"
	"fmt"
	"testing"
)

var errSome = errors.New("hello world")

func TestError_WrapError(t *testing.T) {
	err := E("A", errSome)
	err = E("B", err)
	assertEqual("B: A: hello world", err.Error())
	assertEqual(EInternal, ErrorCode(err))
}

func TestError_TranslateError(t *testing.T) {
	err := E("A", EDuplicate, "duplicate")
	err = E("B", EInternal, err, "fatal error")
	assertEqual("B: A: <500 fatal error> <600 duplicate> ", err.Error())
	assertEqual(EInternal, ErrorCode(err))
	assertEqual("fatal error", ErrorMsg(err))
}

func TestError_ChangeErrorCode(t *testing.T) {
	err := E("A", EInternal, errSome)
	err = E("B", EDuplicate, err)
	assertEqual("B: A: <600 > <500 > hello world", err.Error())
	assertEqual(EDuplicate, ErrorCode(err))
	assertEqual("hello world", ErrorMsg(err))
}

func TestError_ErrorMsg(t *testing.T) {
	err := E("A", EBadRequest)
	assertEqual("An internal error has occurred. Please contact technical support.", ErrorMsg(err))

	err = E("B")
	assertEqual("An internal error has occurred. Please contact technical support.", ErrorMsg(err))

	err = E("B", "hello world")
	assertEqual("hello world", ErrorMsg(err))
}

func assertEqual(expect, actual interface{}) {
	if expect != actual {
		panic(fmt.Sprintf("failed: expect<%v>, actual<%v>", expect, actual))
	}
}
