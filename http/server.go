// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/nil-go/konf"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/nil-go/nilgo/http/internal"
)

const unix = "unix"

// Run wraps start/stop of the HTTP/1 and HTTP/2 clear text server in a single run function
// with listening on multiple tcp and unix socket address.
//
// It also resister built-in interceptors, e.g recovery, log buffering, and timeout.
func Run(server *http.Server, opts ...Option) func(context.Context) error { //nolint:cyclop,funlen,gocognit
	option := &options{}
	for _, opt := range opts {
		opt(option)
	}
	if option.timeout == 0 {
		option.timeout = 10 * time.Second //nolint:gomnd
	}
	if server == nil {
		server = &http.Server{
			ReadTimeout: option.timeout,
		}
	}
	if server.ReadTimeout == 0 {
		// It has to be longer than the timeout of the handler.
		server.ReadTimeout = option.timeout * 2 //nolint:gomnd
	}
	if server.WriteTimeout == 0 {
		server.WriteTimeout = server.ReadTimeout
	}
	if server.IdleTimeout == 0 {
		server.IdleTimeout = server.ReadTimeout * 3 //nolint:gomnd
	}

	if len(option.addresses) == 0 {
		address := "localhost:8080"
		if a := os.Getenv("PORT"); a != "" {
			address = ":" + a
		}
		option.addresses = []socket{{network: "tcp", address: address}}
	}

	handler := server.Handler
	if handler == nil {
		handler = http.DefaultServeMux
	}
	if option.configs != nil {
		mux := http.NewServeMux()
		mux.Handle("/", handler)
		mux.HandleFunc("GET /_config/{path}", config(option))
		handler = mux
	}
	logHandler := slog.Default().Handler()
	handler = internal.RecoveryInterceptor(handler, logHandler)
	if internal.IsSamplingHandler(logHandler) {
		handler = internal.BufferInterceptor(handler)
	}
	handler = http.TimeoutHandler(handler, option.timeout, "request timeout")
	server.Handler = h2c.NewHandler(handler, &http2.Server{})

	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(option.addresses) + 1)
		if slices.ContainsFunc(option.addresses, func(addr socket) bool { return addr.network == unix }) {
			if transport, ok := http.DefaultTransport.(*http.Transport); ok {
				internal.RegisterUnixProtocol(transport)
			}
		}
		for _, addr := range option.addresses {
			addr := addr
			go func() {
				defer waitGroup.Done()

				if addr.network == unix {
					if err := os.RemoveAll(addr.address); err != nil {
						slog.LogAttrs(ctx, slog.LevelWarn, "Could not delete unix socket file.", slog.Any("error", err))
					}
				}
				listener, err := net.Listen(addr.network, addr.address)
				if err != nil {
					cancel(fmt.Errorf("start listener: %w", err))

					return
				}

				slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("HTTP Server listens on %s.", listener.Addr()))
				if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
					cancel(fmt.Errorf("start HTTP Server on %s: %w", listener.Addr(), err))
				}
			}()
		}
		go func() {
			defer waitGroup.Done()

			<-ctx.Done()
			if err := server.Shutdown(context.WithoutCancel(ctx)); err != nil {
				cancel(fmt.Errorf("shutdown HTTP Server: %w", err))
			}
			slog.LogAttrs(ctx, slog.LevelInfo, "HTTP Server is stopped.")
		}()
		waitGroup.Wait()

		if err := context.Cause(ctx); err != nil && !errors.Is(err, ctx.Err()) {
			return err //nolint:wrapcheck
		}

		return nil
	}
}

func config(option *options) func(write http.ResponseWriter, request *http.Request) {
	return func(write http.ResponseWriter, request *http.Request) {
		var err error
		defer func() {
			if err != nil {
				http.Error(write, err.Error(), http.StatusInternalServerError)
			}
		}()

		path := request.PathValue("path")
		if len(option.configs) == 0 {
			_, err = write.Write([]byte(konf.Explain(path)))

			return
		}

		for i, config := range option.configs {
			if i > 0 {
				if _, err := write.Write([]byte("\n-----\n")); err != nil {
					return
				}
			}
			if _, err := write.Write([]byte(config.Explain(path))); err != nil {
				return
			}
		}
	}
}
