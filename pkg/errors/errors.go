package errors

import (
	"context"
	"errors"
	"fmt"
)

type Code int

func (c Code) Int() int {
	return int(c)
}

const (
	_                    Code = iota // Skip the first value of 0
	CodeNotFound                     // CodeNotFound = 1
	CodeStatusConflicted             // CodeStatusConflicted = 2
	CodeInvalidAuth
	CodeBadParameters
	CodeTimeout
	CodeInternal
	CodeNoMoreRetry
	CodeBadGateway
	CodeUnknown
	CodeNotImpl
)

type unwrapper interface {
	Unwrap() error
}

var (
	_ error     = &Error{}
	_ unwrapper = &Error{}
)

type Error struct {
	code Code
	err  error
}

func (f *Error) Unwrap() error {
	if f == nil {
		return nil
	}
	if f.err == nil {
		return nil
	}
	return f.err
}

func (f *Error) Error() string {
	if f == nil {
		return "nil err"
	}
	if f.err == nil {
		return "nil err"
	}
	return fmt.Sprintf("code(%d), %s", f.code.Int(), f.err)
}

func (f *Error) Code() Code {
	return f.code
}

func New(c Code, err error) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		code: c,
		err:  err,
	}
}

func Newf(c Code, format string, args ...interface{}) *Error {
	return &Error{
		code: c,
		err:  fmt.Errorf(format, args...),
	}
}

func As(err error) (bool, *Error) {
	e := &Error{}
	as := errors.As(err, &e)
	if !as {
		return false, nil
	}
	return true, e
}

//func Is(err error) (bool, *Error) {
//	fmt.Printf("err:%+v\n", err)
//	myErr := errors.Is(err, &Error{})
//	if !myErr {
//		return false, nil
//	}
//	internalerr, _ := err.(*Error)
//	return true, internalerr
//}

func NewUnknownError(err error) *Error {
	return New(CodeUnknown, err)
}

func NewStatusConflicted(err error) *Error {
	return New(CodeStatusConflicted, err)
}

func NewInvalidAuth(err error) *Error {
	return New(CodeInvalidAuth, err)
}

func NewTimeoutError(err error) *Error {
	return New(CodeTimeout, err)
}

func NewInternalError(err error) *Error {
	return New(CodeInternal, err)
}

func NewBadParamsError(err error) *Error {
	return New(CodeBadParameters, err)
}

func NewBadGatewayError(err error) *Error {
	return New(CodeBadGateway, err)
}

func NewNoMoreRetryError(err error) *Error {
	return New(CodeNoMoreRetry, err)
}

func NewNotFoundError(err error) *Error {
	return New(CodeNotFound, err)
}

func NewNotImplError(err error) *Error {
	return New(CodeNotImpl, err)
}

func NewContextError(ctx context.Context) *Error {
	select {
	case <-ctx.Done():
		return newContextError(ctx.Err())
	default:
		return newContextError(errors.New("context is not finished"))
	}
}

func newContextError(err error) *Error {
	switch err {
	case nil:
		return nil
	case context.DeadlineExceeded:
		return NewTimeoutError(err)
	case context.Canceled:
		return nil
	default:
		return NewUnknownError(err)
	}
}
