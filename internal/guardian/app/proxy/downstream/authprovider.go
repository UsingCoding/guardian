package downstream

import (
	"github.com/UsingCoding/fpgo/pkg/maybe"

	"guardian/internal/guardian/app/user"
)

type AuthProvider interface {
	AuthorizerByType(t AuthorizerType) maybe.Maybe[Authorizer]
}

type AuthorizerFactoryConfig struct {
	cookie maybe.Maybe[string]
}

func NewAuthorizerFactory(
	config AuthorizerFactoryConfig,
	userProvider user.Provider,
) AuthProvider {
	return &authorizerProvider{
		authorizers: buildAuthorizers(config, userProvider),
	}
}

type authorizerProvider struct {
	authorizers map[AuthorizerType]Authorizer
}

func (provider *authorizerProvider) AuthorizerByType(t AuthorizerType) maybe.Maybe[Authorizer] {
	authorizer, e := provider.authorizers[t]
	if !e {
		return maybe.Maybe[Authorizer]{}
	}
	return maybe.NewJust(authorizer)
}

func buildAuthorizers(
	config AuthorizerFactoryConfig,
	userProvider user.Provider,
) map[AuthorizerType]Authorizer {
	authorizers := map[AuthorizerType]Authorizer{}

	if c, ok := maybe.JustValid(config.cookie); ok {
		authorizers[CookieAuthorizer] = &cookieAuthorizer{
			cookieName:   c,
			userProvider: userProvider,
		}
	}

	return authorizers
}
