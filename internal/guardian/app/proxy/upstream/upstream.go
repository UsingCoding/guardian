package upstream

import (
	"net/url"

	"github.com/UsingCoding/fpgo/pkg/maybe"
)

type AuthorizerType string

const (
	HeaderAuthorizerType = string("header")
)

type Upstream struct {
	ID      string
	Address *url.URL

	Authorizer maybe.Maybe[Authorizer]
}
