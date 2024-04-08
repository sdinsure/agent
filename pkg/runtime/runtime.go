package runtime

import (
	"context"

	"github.com/google/uuid"
)

type RequestIdentityResolver interface {
	// WithRequestID set request id (uuid) into context
	WithRequestID(context.Context) context.Context

	// RequestID retrieves request id from context
	RequestID(context.Context) (TypeRequestID, bool)
}

type (
	requestIdKey struct{}
)

type TypeRequestID string

func (t TypeRequestID) String() string {
	return string(t)
}

type UUIDRequestIdentityResolver struct{}

var (
	_ RequestIdentityResolver = &UUIDRequestIdentityResolver{}
)

func (u *UUIDRequestIdentityResolver) WithRequestID(ctx context.Context) context.Context {
	return context.WithValue(ctx, requestIdKey{}, TypeRequestID(uuid.NewString()))
}

func (u *UUIDRequestIdentityResolver) RequestID(ctx context.Context) (TypeRequestID, bool) {
	info := ctx.Value(requestIdKey{})
	infov, castable := info.(TypeRequestID)
	return infov, castable
}
