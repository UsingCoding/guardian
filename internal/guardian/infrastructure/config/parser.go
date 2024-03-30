package config

import (
	"net/url"
	"os"
	"path"

	"github.com/UsingCoding/fpgo/pkg/maybe"
	"github.com/UsingCoding/fpgo/pkg/slices"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/gohcl"
	"github.com/hashicorp/hcl/v2/hclsimple"
	"github.com/pkg/errors"

	"guardian/internal/guardian/app/config"
	appdownstream "guardian/internal/guardian/app/proxy/downstream"
	appupstream "guardian/internal/guardian/app/proxy/upstream"
	"guardian/internal/guardian/app/user"
	"guardian/internal/guardian/infrastructure/ldap"
)

type Parser struct{}

func (p Parser) Parse(file string) (config.AppConfig, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return config.AppConfig{}, errors.Wrap(err, "failed to read config")
	}
	return p.ParseData(path.Base(file), data)
}

func (p Parser) ParseData(filename string, data []byte) (config.AppConfig, error) {
	var c appConfig
	err := hclsimple.Decode(filename, data, nil, &c)
	if err != nil {
		return config.AppConfig{}, errors.Wrap(err, "failed to parse app config")
	}

	var provider maybe.Maybe[user.Provider]

	if c.UserProvider != nil {
		p, err2 := mapUserProvider(*c.UserProvider)
		if err2 != nil {
			return config.AppConfig{}, err2
		}

		provider = maybe.NewJust(p)
	}

	servers, err := mapHTTPProxies(c.HTTPProxies, provider)
	if err != nil {
		return config.AppConfig{}, err
	}

	return config.AppConfig{
		Healthcheck: config.Healthcheck{
			Address: c.Healthcheck.Address,
			Path:    c.Healthcheck.Path,
		},
		UserProvider: provider,
		TCPProxies:   mapTCPProxy(c.TCPProxies),
		HTTPProxies:  servers,
	}, nil
}

func mapUserProvider(provider userProvider) (user.Provider, error) {
	switch provider.Type {
	case ldapUserProviderType:
		return ldap.NewUserProvider(provider.Address), nil
	default:
		return nil, errors.Errorf("unknown user provider type %s", provider.Type)
	}
}

func mapTCPProxy(proxies []tcpProxy) []config.TCPProxy {
	return slices.Map(proxies, func(p tcpProxy) config.TCPProxy {
		return config.TCPProxy{
			SrcAddress: p.SrcAdress,
			DstAddress: p.DstAddress,
		}
	})
}

func mapHTTPProxies(proxies []httpProxy, provider maybe.Maybe[user.Provider]) ([]config.HTTPProxy, error) {
	return slices.MapErr(proxies, func(s httpProxy) (config.HTTPProxy, error) {
		d, err := mapDownstream(s, provider)
		if err != nil {
			return config.HTTPProxy{}, err
		}

		u, err := mapUpstream(s.Upstream)
		if err != nil {
			return config.HTTPProxy{}, err
		}

		return config.HTTPProxy{
			Address: s.Address,
			Limit: config.Limit{
				RPS:   s.Limit.RPS,
				Burst: s.Limit.Burst,
			},
			Downstream: d,
			Upstream:   u,
		}, nil
	})
}

func mapDownstream(s httpProxy, provider maybe.Maybe[user.Provider]) ([]appdownstream.Downstream, error) {
	return slices.MapErr(s.Downstream, func(d downstream) (appdownstream.Downstream, error) {
		if !maybe.Valid(findUpstream(s.Upstream, d.UpstreamID)) {
			return appdownstream.Downstream{}, errors.Errorf(
				"upstream %s for downstream %s not found",
				d.UpstreamID,
				d.ID,
			)
		}

		rules, err := mapRules(d.Rules)
		if err != nil {
			return appdownstream.Downstream{}, err
		}

		var a maybe.Maybe[appdownstream.Authorizer]

		if d.Authorizer != nil {
			p, ok := maybe.JustValid(provider)
			if !ok {
				return appdownstream.Downstream{}, errors.Errorf("downstream %s requires authorizer but no user provider configured", d.ID)
			}

			auth, err2 := mapDownstreamAuthorizer(*d.Authorizer, p)
			if err2 != nil {
				return appdownstream.Downstream{}, err2
			}

			a = maybe.NewJust(auth)
		}

		return appdownstream.Downstream{
			ID:         d.ID,
			Rules:      rules,
			UpstreamID: d.UpstreamID,
			Authorizer: a,
		}, nil
	})
}

func mapRules(rules []rule) ([]appdownstream.Rule, error) {
	return slices.MapErr(rules, func(r rule) (appdownstream.Rule, error) {
		switch r.Type {
		case hostRuleType:
			decodedRule, err := decodeHclBody[hostRule](r.Payload)
			if err != nil {
				return nil, err
			}

			return appdownstream.HostDownstreamRule{
				Host: decodedRule.Host,
			}, nil
		case pathPrefixRuleType:
			decodedRule, err := decodeHclBody[pathPrefixRule](r.Payload)
			if err != nil {
				return nil, err
			}

			return appdownstream.PathPrefix{
				Prefix: decodedRule.Path,
			}, nil
		default:
			return nil, errors.Errorf("unknown rule type %s", r.Type)
		}
	})
}

func mapDownstreamAuthorizer(authorizer downstreamAuthorizer, provider user.Provider) (appdownstream.Authorizer, error) {
	switch authorizer.Type {
	case cookieDownstreamAuthorizerType:
		auth, err := decodeHclBody[cookieDownstreamAuthorizer](authorizer.Payload)
		if err != nil {
			return nil, err
		}

		return appdownstream.NewCookieAuthorizer(auth.Key, provider), nil
	default:
		return nil, errors.Errorf("unknown donstream authorizer %s", authorizer.Type)
	}
}

func mapUpstream(upstreams []upstream) ([]appupstream.Upstream, error) {
	return slices.MapErr(upstreams, func(u upstream) (appupstream.Upstream, error) {
		var a maybe.Maybe[appupstream.Authorizer]
		if u.Authorizer != nil {
			auth, err := mapUpstreamAuthorizer(*u.Authorizer)
			if err != nil {
				return appupstream.Upstream{}, errors.Wrap(err, "failed to map upstream authorizer")
			}
			a = maybe.NewJust(auth)
		}

		address, err := url.Parse(u.Address)
		if err != nil {
			return appupstream.Upstream{}, errors.Wrapf(err, "failed to parse %s upstream addess", u.Address)
		}

		return appupstream.Upstream{
			ID:         u.ID,
			Address:    address,
			Authorizer: a,
		}, nil
	})
}

func mapUpstreamAuthorizer(authorizer upstreamAuthorizer) (appupstream.Authorizer, error) {
	switch authorizer.Type {
	case headerUpstreamAuthorizerType:
		auth, err := decodeHclBody[headerUpstreamAuthorizer](authorizer.Payload)
		if err != nil {
			return nil, err
		}

		return appupstream.NewAuthHeaderAuthorizer(auth.UserID, auth.Username), nil
	default:
		return nil, errors.Errorf("unknown upstream authorizer %s", authorizer.Type)
	}
}

func decodeHclBody[T any](body hcl.Body) (v T, err error) {
	diags := gohcl.DecodeBody(body, nil, &v)
	if diags.HasErrors() {
		return v, diags
	}
	return v, nil
}

func findUpstream(upstreams []upstream, id string) maybe.Maybe[upstream] {
	for _, u := range upstreams {
		if u.ID == id {
			return maybe.NewJust(u)
		}
	}

	return maybe.Maybe[upstream]{}
}
