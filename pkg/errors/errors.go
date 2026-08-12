package errors

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

// GRPCStatus maps this error's Code onto a gRPC status.
//
// Without it, *Error is an ordinary Go error, so grpc-go classifies every one
// of them as codes.Unknown. The semantic Code is carried in the message text
// and nowhere the transport can see it, which means a caller gets the same
// answer for "you are not authenticated", "that does not exist" and "the
// database is down".
//
// Through grpc-gateway that becomes HTTP 500 for all of them. A missing
// credential is reported to a browser as a server fault:
//
//	{"code":2, "message":"code(3), auth: no valid auth and non-annonymous access"}
//
// where code 2 is Unknown and code(3) is CodeInvalidAuth - the real reason,
// visible only by reading the string. A client cannot act on that: 401 sends
// you to a login screen, 500 makes you retry.
//
// grpc-go looks for this method (status.FromError checks the GRPCStatus
// interface), so implementing it is all that is needed - no call site changes.
func (f *Error) GRPCStatus() *status.Status {
	if f == nil {
		return status.New(codes.OK, "")
	}
	return status.New(f.code.GRPCCode(), f.Error())
}

// GRPCCode is the gRPC code this Code corresponds to.
//
// Kept as a method on Code rather than inline so callers can map without
// constructing an Error, and so the correspondence is readable in one place.
func (c Code) GRPCCode() codes.Code {
	switch c {
	case CodeNotFound:
		return codes.NotFound
	case CodeStatusConflicted:
		return codes.AlreadyExists
	case CodeInvalidAuth:
		// PermissionDenied (403), NOT Unauthenticated (401).
		//
		// This code is raised for BOTH kinds of rejection: the authn
		// middleware uses it when there is no usable credential, and the
		// authz middleware uses it when a known caller is refused. One code
		// cannot mean both 401 and 403, so it has to pick the one that is
		// wrong less often - and telling an authenticated user "log in" is
		// the more misleading of the two, because it sends them round a loop
		// that cannot fix anything.
		//
		// This also matches what the known consumer already does downstream:
		// footprintai/grandturks translates CodeInvalidAuth to
		// PermissionDenied at its middleware edge, with the same reasoning
		// (grandturks#775).
		//
		// Splitting 401 from 403 properly needs a SEPARATE code for "no
		// credential presented"; that is a wider change than this one and
		// would want the middlewares updated to use it.
		return codes.PermissionDenied
	case CodeBadParameters:
		return codes.InvalidArgument
	case CodeTimeout:
		return codes.DeadlineExceeded
	case CodeInternal:
		return codes.Internal
	case CodeNoMoreRetry:
		return codes.Aborted
	case CodeBadGateway:
		return codes.Unavailable
	case CodeNotImpl:
		return codes.Unimplemented
	case CodeUnknown:
		return codes.Unknown
	}
	// A Code this switch has not been taught about. Unknown is the honest
	// answer and preserves today's behaviour for it.
	return codes.Unknown
}
