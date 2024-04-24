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

type AuthZMiddlewareOptioner interface {
	apply(o *AuthZMiddlewareOptions)
}

type AuthZMiddlewareOptions struct {
	skippedAuthPaths []HttpPath
	caseSensitive    bool

	skippedSubjects []string
}

type skippedAuthzPaths struct {
	paths         []HttpPath
	caseSensitive bool
}

func (s skippedAuthzPaths) apply(o *AuthZMiddlewareOptions) {
	o.skippedAuthPaths = append(o.skippedAuthPaths, s.paths...)
	o.caseSensitive = s.caseSensitive
}

func WithSkippedAuthZPaths(paths []HttpPath, caseSensitive bool) AuthZMiddlewareOptioner {
	return skippedAuthzPaths{paths: paths, caseSensitive: caseSensitive}
}

type HttpPath struct {
	RawPath   string
	RawMethod string
}

type skippedSubjects struct {
	subjects []string
}

func (s skippedSubjects) apply(o *AuthZMiddlewareOptions) {
	o.skippedSubjects = append(o.skippedSubjects, s.subjects...)
}

func WithSkippedSubjects(subjects ...string) skippedSubjects {
	return skippedSubjects{subjects: subjects}
}

func newOptions(opts ...AuthZMiddlewareOptioner) *AuthZMiddlewareOptions {
	o := &AuthZMiddlewareOptions{}
	for _, opt := range opts {
		opt.apply(o)
	}
	return o

}

type Enforcer interface {
	// Enforce returns true if the $subject can perform $action on the $object
	Enforce(ctx context.Context, subject, object, action string) (bool, error)
}

func NewAuthZMiddleware(log logger.Logger, enforcer Enforcer, optioners ...AuthZMiddlewareOptioner) *AuthzMiddleware {
	opt := newOptions(optioners...)
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

	sub, found := runtime.SubInfo(ctx)
	if !found {
		return nil, sderrors.NewInvalidAuth(errors.New("invalid sub info"))
	}

	if canSkip, err := a.checkBypass(ctx, sub, httpPath, httpVerb); err != nil {
		return ctx, err
	} else if canSkip {
		a.log.Infox(ctx, "authz: reuqest skipeed")
		return ctx, nil
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

func (a *AuthzMiddleware) checkBypass(ctx context.Context, sub, httpPath, httpMethod string) (bool, error) {

	if a.opt != nil && len(a.opt.skippedSubjects) > 0 {
		a.log.Infox(ctx, "authnmiddleware check skipped sub, len:%d", len(a.opt.skippedSubjects))
		for _, skippedSub := range a.opt.skippedSubjects {
			if sub == skippedSub {
				a.log.Infox(ctx, "authnmiddleware sub(%s) matched, skipped\n", sub)
				return true, nil
			}
		}
	}

	if a.opt != nil && len(a.opt.skippedAuthPaths) > 0 {
		for _, skippedAuthPath := range a.opt.skippedAuthPaths {
			a.log.Infox(ctx, "authnmiddleware http path:%s, method:%s", httpPath, httpMethod)
			a.log.Infox(ctx, "authnmiddleware skipped path:%+v", skippedAuthPath)
			if a.opt.caseSensitive {
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
