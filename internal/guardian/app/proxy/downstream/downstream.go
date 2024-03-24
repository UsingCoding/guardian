package downstream

import (
	"github.com/UsingCoding/fpgo/pkg/maybe"
)

type AuthorizerType string

const (
	CookieAuthorizer = AuthorizerType("cookie-authorizer")
)

type Downstream struct {
	ID string

	Rules []Rule

	UpstreamID string
	Authorizer maybe.Maybe[Authorizer]
}
