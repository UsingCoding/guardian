package httpproxy

import (
	"net/url"
	"time"

	"guardian/internal/common/infrastructure/logger"
)

type proxyLog struct {
	DownstreamURL *url.URL
	UpstreamURL   *url.URL
	Start         time.Time
}

func (p *proxy) logProxy(l proxyLog) {
	p.logger.
		WithFields(transformFields(l)).
		Info("proxy completed")
}

func (p *proxy) logProxyErr(err error, l proxyLog) {
	p.logger.
		WithFields(transformFields(l)).
		Error(err)
}

func transformFields(l proxyLog) logger.Fields {
	return logger.Fields{
		"downstream": l.DownstreamURL,
		"upstream":   l.UpstreamURL,
		"duration":   time.Since(l.Start).String(),
	}
}
