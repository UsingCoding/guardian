package upstream

import (
	"context"
	"net/http"

	"guardian/internal/guardian/app/user"
)

type Authorizer interface {
	Authorize(ctx context.Context, r *http.Request, token user.Descriptor)
}

func NewAuthHeaderAuthorizer(headerID, headerUsername string) Authorizer {
	return &authHeaderAuthorizer{headerID: headerID, headerUsername: headerUsername}
}

type authHeaderAuthorizer struct {
	headerID       string
	headerUsername string
}

func (auth *authHeaderAuthorizer) Authorize(_ context.Context, r *http.Request, descriptor user.Descriptor) {
	r.Header.Set(auth.headerID, descriptor.ID.String())
	r.Header.Set(auth.headerUsername, descriptor.Username)
}
