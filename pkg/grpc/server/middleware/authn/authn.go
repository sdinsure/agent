package authnmiddleware

import (
	"context"

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

func NewAuthNMiddleware(l logger.Logger, claimParser ClaimParser) *AuthNMiddleware {
	return &AuthNMiddleware{
		log:         l,
		tokenParser: &bearerTokenParser{l: l},
		claimParser: claimParser,
	}
}

type AuthNMiddleware struct {
	log         logger.Logger
	tokenParser TokenParser
	claimParser ClaimParser
}

func (a *AuthNMiddleware) AuthFunc(ctx context.Context) (context.Context, error) {
	// allows useclaim to be overwritten, so the priority bearer > api-key
	token, svrerr := a.tokenParser.ParseToken(ctx)
	if svrerr != nil || len(token) == 0 {
		a.log.Errorx(ctx, "no auth token found, possible err:%+v\n", svrerr)
		return ctx, svrerr
	}
	userClaim, err := a.claimParser.ParseClaim(ctx, token)
	if err != nil {
		a.log.Error("auth: invalid auth tokne:%v\n", err)
		return nil, sderrors.NewInvalidAuth(err)
	}
	sub, _ := userClaim.GetSubject()
	a.log.Info("auth: populate relavant info to ctx\n")
	return runtime.WithKeyInfo(
		runtime.WithClaimInfo(
			runtime.WithSubInfo(runtime.WithRequestId(ctx), sub),
			userClaim,
		),
		token,
	), nil
}
