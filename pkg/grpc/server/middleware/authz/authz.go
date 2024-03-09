package authzmiddleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	sderrors "github.com/sdinsure/agent/pkg/errors"
	"github.com/sdinsure/agent/pkg/grpc/server/runtime"
	"github.com/sdinsure/agent/pkg/logger"
)

type authZMiddlewareOptions interface {
	apply(o *AuthZMiddlewareOptions)
}

type AuthZMiddlewareOptions struct {
	SkippedAuthPaths []HttpPath
	CaseSensitive    bool
}

type AuthZMiddlewareOptionFunc func(o *AuthZMiddlewareOptions)

func (a AuthZMiddlewareOptionFunc) apply(o *AuthZMiddlewareOptions) {
	a(o)
}

func newOptions(opts ...AuthZMiddlewareOptionFunc) *AuthZMiddlewareOptions {
	o := &AuthZMiddlewareOptions{}
	for _, opt := range opts {
		opt(o)
	}
	return o

}

type HttpPath struct {
	RawPath   string
	RawMethod string
}

func WithSkippedAuthZPaths(paths []HttpPath) AuthZMiddlewareOptionFunc {
	return func(o *AuthZMiddlewareOptions) {
		o.SkippedAuthPaths = paths
	}
}

type Enforcer interface {
	// Enforce returns true if the $subject can perform $action on the $object
	Enforce(ctx context.Context, subject, object, action string) (bool, error)
}

func NewAuthZMiddleware(log logger.Logger, enforcer Enforcer, optFuncs ...AuthZMiddlewareOptionFunc) *AuthzMiddleware {
	opt := newOptions(optFuncs...)
	return &AuthzMiddleware{log: log, enforcer: enforcer, opt: opt}
}

type AuthzMiddleware struct {
	log      logger.Logger
	enforcer Enforcer
	opt      *AuthZMiddlewareOptions
}

func (a *AuthzMiddleware) AuthFunc(ctx context.Context) (context.Context, error) {
	a.log.Infox(ctx, "authz: reuqest")
	httpPath, found := runtime.HttpPath(ctx)
	a.log.Info("authz: reuqested path:%+v, found:%+v", httpPath, found)
	if !found {
		return nil, sderrors.NewInvalidAuth(errors.New("invalid http path info"))
	}
	a.log.Info("authz: reuqested path:%+v", httpPath)
	httpVerb, found := runtime.HttpVerb(ctx)
	if !found {
		return nil, sderrors.NewInvalidAuth(errors.New("invalid http verb info"))
	}
	a.log.Info("authz: reuqested httpverb:%+v", httpVerb)

	if canSkip, err := a.checkBypassMethodOrPath(ctx, httpPath, httpVerb); err != nil {
		return ctx, err
	} else if canSkip {
		a.log.Infox(ctx, "authz: reuqest skipeed")
		return ctx, nil
	}

	sub, found := runtime.SubInfo(ctx)
	if !found {
		return nil, sderrors.NewInvalidAuth(errors.New("invalid sub info"))
	}
	a.log.Info("authz: reuqested sub:%+v", sub)

	canPass, err := a.enforcer.Enforce(ctx, sub, httpPath, httpVerb)
	if err != nil {
		return nil, sderrors.NewInvalidAuth(fmt.Errorf("enforce failed, err:%+v\n", err))
	}
	a.log.Info("authz: sub:%s, path:%s, verb:%s, pass?:%+v\n", sub, httpPath, httpVerb, canPass)
	if !canPass {
		return nil, sderrors.NewInvalidAuth(errors.New("permission denied"))
	}
	return ctx, nil
}

func (a *AuthzMiddleware) checkBypassMethodOrPath(ctx context.Context, httpPath, httpMethod string) (bool, error) {
	if a.opt != nil && len(a.opt.SkippedAuthPaths) > 0 {
		for _, skippedAuthPath := range a.opt.SkippedAuthPaths {
			a.log.Infox(ctx, "authnmiddleware http path:%s, method:%s", httpPath, httpMethod)
			a.log.Infox(ctx, "authnmiddleware skipped path:%+v", skippedAuthPath)
			if a.opt.CaseSensitive {
				if skippedAuthPath.RawPath == httpPath && strings.ToLower(skippedAuthPath.RawMethod) == strings.ToLower(httpMethod) {
					a.log.Infox(ctx, "authmiddleware skipped")
					return true, nil
				}
			} else {
				if strings.ToLower(skippedAuthPath.RawPath) == strings.ToLower(httpPath) && strings.ToLower(skippedAuthPath.RawMethod) == strings.ToLower(httpMethod) {
					a.log.Infox(ctx, "authmiddleware skipped")
					return true, nil
				}
			}
		}
	}
	return false, nil
}
