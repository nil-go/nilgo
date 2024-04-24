// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"sync"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"

	"github.com/nil-go/nilgo/grpc/log"
	pb "github.com/nil-go/nilgo/grpc/pb/nilgo/v1"
)

// NewServer creates a new gRPC server with the given options.
//
// It wraps grpc.NewServer with built-in interceptors, e.g recovery, log buffering, and statsHandler.
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	option := &serverOptions{}
	for _, opt := range opts {
		switch so := opt.(type) {
		case serverOptionFunc:
			so.fn(option)
		default:
			option.grpcOpts = append(option.grpcOpts, opt)
		}
	}

	builtInOpts := log.ServerOptions(option.handler)
	builtInOpts = append(builtInOpts, grpc.WaitForHandlers(true))
	if option.otelOpts != nil {
		builtInOpts = append(builtInOpts, grpc.StatsHandler(otelgrpc.NewServerHandler(option.otelOpts...)))
	}
	server := grpc.NewServer(append(builtInOpts, option.grpcOpts...)...)
	if option.configs != nil {
		pb.RegisterConfigServiceServer(server, &ConfigServiceServer{configs: option.configs})
	}

	return server
}

// Run wraps start/stop of the gRPC server in a single run function
// with listening on multiple tcp and unix socket address.
//
// It also resister health and reflection services if the services have not registered.
func Run(server *grpc.Server, addresses ...string) func(context.Context) error { //nolint:cyclop,funlen,gocognit
	if server == nil {
		server = grpc.NewServer()
	}
	if len(addresses) == 0 {
		address := "localhost:8080"
		if a := os.Getenv("PORT"); a != "" {
			address = ":" + a
		}
		addresses = []string{address}
	}

	// Register health service if necessary.
	var healthServer *health.Server
	if _, exist := server.GetServiceInfo()[grpc_health_v1.Health_ServiceDesc.ServiceName]; !exist {
		healthServer = health.NewServer()
		defer healthServer.Resume()
		grpc_health_v1.RegisterHealthServer(server, healthServer)
	}
	// Register reflection service if necessary.
	if _, exist := server.GetServiceInfo()[grpc_reflection_v1.ServerReflection_ServiceDesc.ServiceName]; !exist {
		reflection.Register(server)
	}

	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)

		var waitGroup sync.WaitGroup
		waitGroup.Add(len(addresses) + 1)
		for _, addr := range addresses {
			addr := addr
			go func() {
				defer waitGroup.Done()

				network := "tcp"
				if strings.HasPrefix(addr, "unix:") {
					network = "unix"
					addr = strings.TrimPrefix(addr[5:], "//")

					if err := os.RemoveAll(addr); err != nil {
						slog.LogAttrs(ctx, slog.LevelWarn, "Could not delete unix socket file.", slog.Any("error", err))
					}
				}
				listener, err := net.Listen(network, addr)
				if err != nil {
					cancel(fmt.Errorf("start listener: %w", err))

					return
				}

				slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("gRPC Server listens on %s.", listener.Addr()))
				if err := server.Serve(listener); err != nil {
					cancel(fmt.Errorf("start gRPC Server on %s: %w", listener.Addr(), err))
				}
			}()
		}
		go func() {
			defer waitGroup.Done()

			<-ctx.Done()
			if healthServer != nil {
				// Shutdown health server so client knows it's not serving.
				healthServer.Shutdown()
			}
			server.GracefulStop()
			slog.LogAttrs(ctx, slog.LevelInfo, "gRPC Server is stopped.")
		}()
		waitGroup.Wait()

		if err := context.Cause(ctx); err != nil && !errors.Is(err, ctx.Err()) {
			return err //nolint:wrapcheck
		}

		return nil
	}
}
