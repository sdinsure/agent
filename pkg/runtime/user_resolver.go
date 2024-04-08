package runtime

import (
	"context"

	"github.com/sdinsure/agent/pkg/grpc/server/runtime"
	"github.com/sdinsure/agent/pkg/logger"
)

type UserResolver interface {
	// WithUserInfo set user information into context
	WithUserInfo(ctx context.Context) context.Context

	// UserInfo retrieves userinfo from context
	UserInfo(ctx context.Context) (UserInfor, error)
}

type IdentityResolver struct {
	log        logger.Logger
	userGetter UserGetter
}

type UserGetter interface {
	GetUser(ctx context.Context, userSub string) (UserInfor, error)
}

var (
	_           pkgruntime.UserResolver = &IdentityResolver{}
	userInfoKey struct{}
)

func NewIdentityResolver(log logger.Logger, userGetter UserGetter) *IdentityResolver {
	return &IdentityResolver{
		log:        log,
		userGetter: userGetter,
	}
}

func (i *IdentityResolver) WithUserInfo(ctx context.Context) context.Context {
	i.log.Infox(ctx, "identityresolver, with user info is called\n")
	sub, hasSub := runtime.SubInfo(ctx)
	if !hasSub {
		return context.WithValue(ctx, userInfoKey{}, annonymous)
	}
	i.log.Infox(ctx, "identityresolver, sub:%+v\n", sub)

	userInfo, err := i.userGetter(ctx, sub)
	if err != nil {
		i.log.Errorx(ctx, "failed to retrieve userinfo, sub:%+v, err:%+v\n", sub, err)
		return context.WithValue(ctx, userInfoKey{}, annonymous)
	}
	i.log.Infox(ctx, "identityresolver, userInfo:%+v\n", userInfo)
	return context.WithValue(ctx, userInfoKey{}, userInfo)
}

func (i *IdentityResolver) UserInfo(ctx context.Context) (UserInfor, bool) {
	info := ctx.Value(userInfoKey{})
	infov, castable := info.(UserInfor)
	return infov, castable
}

type TypeUserId string

func NewTypeUserId(s string) TypeUserId {
	return TypeUserId(s)
}

type TypeUserEmail string

func NewTypeUserEmail(e string) TypeUserEmail {
	return TypeUserEmail(e)
}

type TypeUserGroups []string

func NewTypeUserGroups(gList []string) TypeUserGroups {
	return TypeUserGroups(gList)
}

var (
	_ UserInfor = userInfo{}
)

type UserInfor interface {
	GetUserId() TypeUserId
	GetEmail() TypeUserEmail
	GetGroups() TypeUserGroups
}

type userInfo struct {
	uid    TypeUserID
	email  TypeUserEmail
	groups TypeUserGroups
}

func (u userInfo) GetUserId() TypeUserID {
	return u.uid
}

func (u userInfo) GetEmail() TypeUserEmail {
	return u.email
}

func (u userInfo) GetGroups() TypeUserGroups {
	return u.groups
}

var (
	annonymous = userInfo{
		uid:    NewTypeUserId("annonymous"),
		email:  NewTypeUserEmail("annonymous"),
		groups: NewTypeUserGroups([]string{"annonymous"}),
	}
)
