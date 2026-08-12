package errors

import (
	"errors"
	"fmt"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestErrorCarriesItsGRPCStatus is the behaviour that was missing: grpc-go
// asks an error for its status via the GRPCStatus interface, and without it
// every *Error was classified codes.Unknown - which grpc-gateway renders as
// HTTP 500 regardless of what actually went wrong.
func TestErrorCarriesItsGRPCStatus(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want codes.Code
	}{
		// The case that prompted this: a request with no credential was
		// reported to clients as a server fault instead of 401.
		// CodeInvalidAuth covers BOTH "no credential" and "known caller
		// refused", so it maps to PermissionDenied - see the switch for why
		// one code cannot be both 401 and 403.
		{"invalid auth", NewInvalidAuth(errors.New("no valid auth")), codes.PermissionDenied},
		{"not found", New(CodeNotFound, errors.New("nope")), codes.NotFound},
		{"conflict", NewStatusConflicted(errors.New("exists")), codes.AlreadyExists},
		{"bad parameters", NewBadParamsError(errors.New("bad")), codes.InvalidArgument},
		{"timeout", NewTimeoutError(errors.New("slow")), codes.DeadlineExceeded},
		{"internal", NewInternalError(errors.New("boom")), codes.Internal},
		{"bad gateway", NewBadGatewayError(errors.New("upstream")), codes.Unavailable},
		{"not implemented", New(CodeNotImpl, errors.New("todo")), codes.Unimplemented},
		{"unknown stays unknown", NewUnknownError(errors.New("?")), codes.Unknown},
	} {
		t.Run(tc.name, func(t *testing.T) {
			st, ok := status.FromError(tc.err)
			if !ok {
				t.Fatalf("status.FromError did not recognise the error")
			}
			if st.Code() != tc.want {
				t.Errorf("code = %v, want %v", st.Code(), tc.want)
			}
			// The original text must survive - callers and logs match on it.
			if st.Message() != tc.err.Error() {
				t.Errorf("message = %q, want %q", st.Message(), tc.err.Error())
			}
		})
	}
}

// A nil *Error must not panic when grpc-go asks it for a status.
func TestNilErrorStatusIsOK(t *testing.T) {
	var e *Error
	if got := e.GRPCStatus().Code(); got != codes.OK {
		t.Errorf("nil error code = %v, want OK", got)
	}
}

// Wrapped errors still resolve, since grpc-go unwraps looking for the
// interface - this is how the middleware's errors reach the transport.
func TestWrappedErrorKeepsItsCode(t *testing.T) {
	wrapped := fmt.Errorf("middleware: %w", NewInvalidAuth(errors.New("no valid auth")))
	st, ok := status.FromError(wrapped)
	if !ok {
		t.Fatal("status.FromError did not recognise the wrapped error")
	}
	if st.Code() != codes.PermissionDenied {
		t.Errorf("wrapped code = %v, want PermissionDenied", st.Code())
	}
}
