package runtime

import (
	"context"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type subInfo struct{}
type claimInfo struct{}
type keyInfo struct{}
type requestId struct{}

func WithSubInfo(ctx context.Context, subject string) context.Context {
	return context.WithValue(ctx, subInfo{}, subject)
}

func SubInfo(ctx context.Context) (string, bool) {
	info := ctx.Value(subInfo{})
	if _, castable := info.(string); castable {
		return info.(string), true
	}
	return "", false
}

func WithClaimInfo(ctx context.Context, claims jwt.Claims) context.Context {
	return context.WithValue(ctx, claimInfo{}, claims)
}

func ClaimInfo(ctx context.Context) (jwt.Claims, bool) {
	info := ctx.Value(claimInfo{})
	if _, castable := info.(jwt.Claims); castable {
		return info.(jwt.Claims), true
	}
	return jwt.RegisteredClaims{}, false
}

func WithKeyInfo(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, keyInfo{}, key)
}

func KeyInfo(ctx context.Context) (string, bool) {
	info := ctx.Value(keyInfo{})
	if _, castable := info.(string); castable {
		return info.(string), true
	}
	return "", false
}

func WithRequestId(ctx context.Context) context.Context {
	reqId := uuid.NewString()
	return context.WithValue(ctx, requestId{}, reqId)
}

func RequestId(ctx context.Context) (string, bool) {
	info := ctx.Value(requestId{})
	if _, castable := info.(string); castable {
		return info.(string), true
	}
	return "", false
}
