package downstream

import (
	"context"
	"net/http"
	"strings"
)

type Rule interface {
	Match(ctx context.Context, r http.Request) bool
}

type HostDownstreamRule struct {
	Host string
}

func (h HostDownstreamRule) Match(_ context.Context, r http.Request) bool {
	return r.Host == h.Host
}

type PathPrefix struct {
	Prefix string
}

func (p PathPrefix) Match(_ context.Context, r http.Request) bool {
	return strings.HasPrefix(r.URL.Path, p.Prefix)
}
