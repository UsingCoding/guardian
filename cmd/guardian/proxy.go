package main

import (
	"io"
	"os"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	commonserver "guardian/internal/common/infrastructure/server"
	"guardian/internal/common/proc"
	"guardian/internal/guardian/app/config"
	infraconfig "guardian/internal/guardian/infrastructure/config"
	infraproxy "guardian/internal/guardian/infrastructure/httpproxy"
	"guardian/internal/guardian/infrastructure/tcpproxy"
)

func proxy() *cli.Command {
	return &cli.Command{
		Name:   "proxy",
		Usage:  "Runs proxy",
		Action: executeProxy,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "config",
				Aliases: []string{"c"},
				Usage:   "Path to config file",
				EnvVars: []string{"GUARDIAN_CONFIG"},
			},
		},
	}
}

func executeProxy(ctx *cli.Context) error {
	l := initLogger()

	configPath := ctx.String("config")

	c, err := loadConfig(configPath)
	if err != nil {
		return err
	}

	hub := proc.NewHub(ctx.Context)

	router := mux.NewRouter()
	commonserver.RegisterHealthCheck(router, c.Healthcheck.Path)
	httpServer(
		hub,
		c.Healthcheck.Address,
		router,
	)

	for _, server := range c.HTTPProxies {
		p := infraproxy.NewProxy(
			server.Downstream,
			server.Upstream,
			server.Limit,
			l,
		)

		httpServer(
			hub,
			server.Address,
			p.Proxy(),
		)
	}

	if len(c.TCPProxies) != 0 {
		p := tcpproxy.Proxy{}
		hub.AddProc(proc.NewProc(
			func() error {
				return p.Proxy(ctx.Context, c.TCPProxies)
			},
			func() error {
				return nil
			},
		))
	}

	return hub.Wait()
}

func loadConfig(p string) (config.AppConfig, error) {
	if p != "" {
		return infraconfig.Parser{}.Parse(p)
	}

	data, err := io.ReadAll(os.Stdin)
	if err != nil && err != io.EOF {
		return config.AppConfig{}, errors.Wrap(err, "failed to read stdin")
	}

	if len(data) == 0 {
		return config.AppConfig{}, errors.New("empty stdin")
	}

	c, err := infraconfig.Parser{}.ParseData("stdin.hcl", data)
	return c, err
}
