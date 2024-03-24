package downstream

import (
	"context"
	stderrors "errors"
	"net/http"

	"github.com/UsingCoding/fpgo/pkg/maybe"
	"github.com/gofrs/uuid/v5"
	"github.com/pkg/errors"

	"guardian/internal/guardian/app/user"
)

var (
	ErrAuthDataNotFound = stderrors.New("auth data not found")
	ErrAuthDataInvalid  = stderrors.New("auth data invalid")
)

type Authorizer interface {
	Auth(ctx context.Context, r http.Request) (user.Descriptor, error)
}

func NewCookieAuthorizer(cookieName string, userProvider user.Provider) Authorizer {
	return &cookieAuthorizer{cookieName: cookieName, userProvider: userProvider}
}

type cookieAuthorizer struct {
	cookieName string

	userProvider user.Provider
}

func (a *cookieAuthorizer) Auth(ctx context.Context, r http.Request) (user.Descriptor, error) {
	var c maybe.Maybe[*http.Cookie]
	for _, cookie := range r.Cookies() {
		if cookie.Name == a.cookieName {
			c = maybe.NewJust(cookie)
		}
	}

	cook, ok := maybe.JustValid(c)
	if !ok {
		return user.Descriptor{}, errors.WithStack(ErrAuthDataNotFound)
	}

	userIDStr := cook.Value
	if userIDStr == "" {
		return user.Descriptor{}, errors.WithStack(ErrAuthDataNotFound)
	}

	userID, err := uuid.FromString(userIDStr)
	if err != nil {
		return user.Descriptor{}, errors.Wrapf(
			ErrAuthDataInvalid,
			"err: %s: userID %s",
			err.Error(),
			userIDStr,
		)
	}

	descriptor, err := a.userProvider.User(ctx, user.Token{
		ID: userID,
	})
	return descriptor, errors.WithStack(err)
}
