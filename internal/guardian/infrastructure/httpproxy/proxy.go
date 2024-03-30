package httpproxy

import (
	"context"
	stderrors "errors"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/UsingCoding/fpgo/pkg/maybe"
	"github.com/pkg/errors"
	"golang.org/x/time/rate"

	"guardian/internal/common/infrastructure/logger"
	"guardian/internal/guardian/app/config"
	"guardian/internal/guardian/app/proxy/downstream"
	"guardian/internal/guardian/app/proxy/upstream"
	"guardian/internal/guardian/app/user"
)

var (
	ErrRequestNotMatched = stderrors.New("request not matched")
	ErrUpstreamNotFound  = stderrors.New("upstream not found")
)

func NewProxy(
	d []downstream.Downstream,
	u []upstream.Upstream,
	limit config.Limit,
	l logger.Logger,
) Proxy {
	var limiter *rate.Limiter
	if limit.RPS != 0 && limit.Burst != 0 {
		limiter = rate.NewLimiter(rate.Limit(limit.RPS), limit.Burst)
	}

	return &proxy{
		d:       d,
		u:       u,
		limiter: limiter,
		logger:  l,
	}
}

type Proxy interface {
	Proxy() http.Handler
}

type proxy struct {
	d []downstream.Downstream
	u []upstream.Upstream

	limiter *rate.Limiter
	logger  logger.Logger
}

type transport struct {
	http.RoundTripper
}

func (t *transport) RoundTrip(request *http.Request) (*http.Response, error) {
	return t.RoundTripper.RoundTrip(request)
}

func (p *proxy) Proxy() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		if p.limiter != nil && !p.limiter.Allow() {
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		res, err := p.proceedRequest(r.Context(), *r)
		if err != nil {
			p.handleErr(err, w, proxyLog{
				DownstreamURL: r.URL,
				UpstreamURL:   nil,
				Start:         start,
			})
			return
		}

		revProxy := &httputil.ReverseProxy{
			Rewrite: func(proxyReq *httputil.ProxyRequest) {
				// set X-Forwarded headers
				proxyReq.SetXForwarded()

				res.ProxyRequestModifier(proxyReq.Out)

				proxyReq.SetURL(res.URL)
				// replace host
				proxyReq.Out.Host = proxyReq.In.Host
			},
			ModifyResponse: func(resp *http.Response) error {
				return res.ResponseReceiver(resp)
			},
			Transport: &transport{
				http.DefaultTransport,
			},
		}

		revProxy.ServeHTTP(w, r)
		p.logProxy(proxyLog{
			DownstreamURL: r.URL,
			UpstreamURL:   res.URL,
			Start:         start,
		})
	})
}

type proceedRes struct {
	URL                  *url.URL
	ProxyRequestModifier func(*http.Request)
	ResponseReceiver     func(*http.Response) error
}

func (p *proxy) proceedRequest(ctx context.Context, r http.Request) (proceedRes, error) {
	d, ok := maybe.JustValid(p.matchDownstream(ctx, r))
	if !ok {
		return proceedRes{}, errors.WithStack(ErrRequestNotMatched)
	}

	u, ok := maybe.JustValid(p.matchUpstream(d.UpstreamID))
	if !ok {
		return proceedRes{}, errors.WithStack(ErrUpstreamNotFound)
	}

	var descriptor maybe.Maybe[user.Descriptor]
	if auth, ok := maybe.JustValid(d.Authorizer); ok {
		desc, err := auth.Auth(ctx, r)
		if err != nil {
			return proceedRes{}, err
		}
		descriptor = maybe.NewJust(desc)
	}

	var authorizer maybe.Maybe[upstream.Authorizer]
	if a, ok := maybe.JustValid(u.Authorizer); ok {
		if !maybe.Valid(descriptor) {
			return proceedRes{}, &ErrUnauthorized{
				Reason: "no user for authorized zone",
			}
		}

		authorizer = maybe.NewJust(a)
	}

	return proceedRes{
		URL: u.Address,
		ProxyRequestModifier: func(request *http.Request) {
			if a, ok := maybe.JustValid(authorizer); ok {
				a.Authorize(
					request.Context(),
					request,
					maybe.Just(descriptor),
				)
			}
		},
		ResponseReceiver: func(_ *http.Response) error {
			return nil
		},
	}, nil
}

func (p *proxy) matchDownstream(ctx context.Context, r http.Request) maybe.Maybe[downstream.Downstream] {
	for _, d := range p.d {
		match := true
		for _, rule := range d.Rules {
			if !rule.Match(ctx, r) {
				match = false
				break
			}
		}

		if match {
			return maybe.NewJust(d)
		}
	}

	return maybe.Maybe[downstream.Downstream]{}
}

func (p *proxy) matchUpstream(upstreamID string) maybe.Maybe[upstream.Upstream] {
	for _, u := range p.u {
		if u.ID == upstreamID {
			return maybe.NewJust(u)
		}
	}

	return maybe.Maybe[upstream.Upstream]{}
}
