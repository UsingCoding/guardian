package tcpproxy

import (
	"context"
	stderrors "errors"

	"github.com/inetaf/tcpproxy"

	"guardian/internal/guardian/app/config"
)

type Proxy struct{}

func (proxy Proxy) Proxy(ctx context.Context, proxies []config.TCPProxy) (err error) {
	var p tcpproxy.Proxy

	for _, tcpProxy := range proxies {
		p.AddRoute(tcpProxy.SrcAddress, tcpproxy.To(tcpProxy.DstAddress))
	}

	go func() {
		err = p.Run()
	}()

	<-ctx.Done()

	return stderrors.Join(err, p.Close())
}
