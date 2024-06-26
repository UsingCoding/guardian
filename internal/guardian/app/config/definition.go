package config

import (
	"github.com/UsingCoding/fpgo/pkg/maybe"

	"guardian/internal/guardian/app/proxy/downstream"
	"guardian/internal/guardian/app/proxy/upstream"
	"guardian/internal/guardian/app/user"
)

type AppConfig struct {
	Healthcheck  Healthcheck
	UserProvider maybe.Maybe[user.Provider]
	TCPProxies   []TCPProxy
	HTTPProxies  []HTTPProxy
}

type TCPProxy struct {
	SrcAddress string
	DstAddress string
}

type Healthcheck struct {
	Address string
	Path    string
}

type HTTPProxy struct {
	Address string

	Limit Limit

	Downstream []downstream.Downstream
	Upstream   []upstream.Upstream
}

type Limit struct {
	RPS   int
	Burst int
}
