package main

import (
	"context"
	stdlog "log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"

	"guardian/internal/common/infrastructure/logger"
	"guardian/internal/common/proc"
)

const (
	appID = "guardian"
)

func main() {
	ctx := context.Background()

	ctx = subscribeForKillSignals(ctx)

	err := runApp(ctx, os.Args)

	switch errors.Cause(err) {
	case proc.ErrStopped:
		stdlog.Println(err.Error())
		return
	default:
		stdlog.Fatal(err)
	}
}

func runApp(ctx context.Context, args []string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	app := &cli.App{
		Name: appID,
		Commands: []*cli.Command{
			proxy(),
		},
	}

	return app.RunContext(ctx, args)
}

func subscribeForKillSignals(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		defer cancel()
		select {
		case <-ctx.Done():
			signal.Stop(ch)
		case <-ch:
		}
	}()

	return ctx
}

func initLogger() logger.Logger {
	return logger.NewLogger(logger.Config{
		AppID: appID,
	})
}

func httpServer(
	hub proc.Hub,
	serveAddr string,
	handler http.Handler,
) {
	var server *http.Server

	hub.AddProc(proc.NewProc(
		func() error {
			server = &http.Server{
				Handler:      handler,
				Addr:         serveAddr,
				WriteTimeout: 15 * time.Second,
				ReadTimeout:  15 * time.Second,
			}

			return server.ListenAndServe()
		},
		func() error {
			if server == nil {
				return nil
			}
			return server.Shutdown(context.Background())
		},
	))
}
