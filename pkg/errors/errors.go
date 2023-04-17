package errors

import (
	goerrors "errors"

	"github.com/pkg/errors"
)

// New returns an error with the supplied message.
// New also records the stack trace at the point it was called.
// just like buildin   errors.New().
func New(message string) error {
	return goerrors.New(message)
}

// Errorf formats according to a format specifier and returns the string
// as a value that satisfies error.
// Errorf also records the stack trace at the point it was called.
// just like  fmt.Errorf().
func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

// WithMessage annotates err with a new message.
// If err is nil, WithMessage returns nil.
// extent err  with more information but not break the origin error.
func WithMessage(err error, message string) error {
	return errors.WithMessage(err, message)
}

// WithMessagef annotates err with the format specifier.
// extent err  with more information but not break the origin error.
func WithMessagef(err error, format string, args ...interface{}) error {
	return errors.WithMessagef(err, format, args...)
}

// Wrap returns an error annotating err with a stack trace
// at the point Wrap is called, and the supplied message.
// If err is nil, Wrap returns nil.
// always used in the boundary of thirdparty module, do not used thirdparty error directory
// but  define a error (var GitError = errors.New("Git Error")) and Warp with the information of third party error info
// Wrap(GitError, error.Error()).
func Wrap(err error, message string) error {
	return errors.Wrap(err, message)
}

// Wrapf returns an error annotating err with a stack trace
// at the point Wrapf is called, and the format specifier.
// If err is nil, Wrapf returns nil.
func Wrapf(err error, format string, args ...interface{}) error {
	return errors.Wrapf(err, format, args...)
}

// WithStack annotates err with a stack trace at the point WithStack was called.
// If err is nil, WithStack returns nil.
func WithStack(err error) error {
	return errors.WithStack(err)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements the following
// interface:
//
//     type causer interface {
//            Cause() error
//     }
//
// If the error does not implement Cause, the original error will
// be returned. If the error is nil, nil will be returned without further
// investigation.

// find the root cause error of the error chain.
func Cause(err error) error {
	return errors.Cause(err)
}
