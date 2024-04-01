package authnmiddleware

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"

	sderrors "github.com/sdinsure/agent/pkg/errors"
	"github.com/sdinsure/agent/pkg/grpc/server/runtime"
	"github.com/sdinsure/agent/pkg/logger"
)

// TokenParser parse auth token from context
type TokenParser interface {
	ParseToken(ctx context.Context) (string, error)
}

type bearerTokenParser struct {
	l logger.Logger
}

func (b *bearerTokenParser) ParseToken(ctx context.Context) (string, error) {
	token, err := grpc_auth.AuthFromMD(ctx, "bearer")
	if err != nil {
		b.l.Error("bearer: parser failed, err:%+v\n", err)
		return "", err
	}
	return token, nil
}

type ClaimParser interface {
	ParseClaim(ctx context.Context, token string) (jwt.Claims, error)
}

type Optional interface {
	apply(*Option)
}

type enableAnnonymous struct {
	enabled bool
}

func (e enableAnnonymous) apply(o *Option) {
	o.enabledAnnonymous = e.enabled
}

func EnableAnnonymous(enabled bool) Optional {
	return enableAnnonymous{enabled: enabled}
}

type Option struct {
	enabledAnnonymous bool
}

func newOption(opts ...Optional) *Option {
	o := &Option{}

	for _, opt := range opts {
		opt.apply(o)
	}
	return o
}

func NewAuthNMiddleware(l logger.Logger, claimParser ClaimParser, option ...Optional) *AuthNMiddleware {

	option := newOption(options...)
	return &AuthNMiddleware{
		log:         l,
		tokenParser: &bearerTokenParser{l: l},
		claimParser: claimParser,
		option:      option,
	}
}

type AuthNMiddleware struct {
	log         logger.Logger
	tokenParser TokenParser
	claimParser ClaimParser
}

func (a *AuthNMiddleware) AuthFunc(ctx context.Context) (context.Context, error) {
	// allows useclaim to be overwritten, so the priority bearer > api-key
	var claims jwt.Claims
	var err error

	token, svrerr := a.tokenParser.ParseToken(ctx)
	if svrerr != nil || len(token) == 0 {
		a.log.Errorx(ctx, "no auth token found, possible err:%+v\n", svrerr)
	}

	if len(token) > 0 {
		a.log.Info(ctx, "token found, parsing its claim\n")
		claims, err = a.claimParser.ParseClaim(ctx, token)
		if err != nil {
			a.log.Error("auth: invalid auth tokne:%v\n", err)
			return nil, sderrors.NewInvalidAuth(err)
		}
	} else if a.option.enabledAnnonymous {
		// if no token found, check whether it is annonmyous
		claims = annonymous{}
	}

	sub, _ := userClaim.GetSubject()
	a.log.Info("auth: populate relavant info to ctx\n")
	return runtime.WithKeyInfo(
		runtime.WithClaimInfo(
			runtime.WithSubInfo(runtime.WithRequestId(ctx), sub),
			claims,
		),
		token,
	), nil
}

// annonymous implements jwt.Claims interface
type annonymous struct{}

var (
	_ jwt.Claims = annonymous{}
)

func (a annonymous) GetExpirationTime() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now().Add(24 * time.Hour)), nil
}

func (a annonymous) GetIssuedAt() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now()), nil
}

func (a annonymous) GetNotBefore() (*jwt.NumericDate, error) {
	return jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)), nil
}

func (a annonymous) GetIssuer() (string, error) {
	return "annonymous", nil
}

func (a annonymous) GetSubject() (string, error) {
	return "annonymous", nil

}
func (a annonymous) GetAudience() (jwt.ClaimStrings, error) {
	return jwt.ClaimStrings([]string{"annonymous"}), nil
}
