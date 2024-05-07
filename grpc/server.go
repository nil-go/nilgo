// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package grpc provides opinionated production-ready gRPC server.
//
// It also redirect gRPC log to slog using init function.
package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"

	"github.com/nil-go/nilgo/grpc/internal"
	pb "github.com/nil-go/nilgo/grpc/pb/nilgo/v1"
)

// NewServer creates a new gRPC server with the given options.
//
// It wraps grpc.NewServer with built-in interceptors, e.g recovery, log buffering.
func NewServer(opts ...grpc.ServerOption) *grpc.Server {
	handler := slog.Default().Handler()
	builtInOpts := []grpc.ServerOption{grpc.WaitForHandlers(true)}
	if internal.IsSamplingHandler(handler) {
		builtInOpts = append(builtInOpts,
			grpc.ChainUnaryInterceptor(internal.BufferUnaryInterceptor),
			grpc.ChainStreamInterceptor(internal.BufferStreamInterceptor),
		)
	}
	// Add recovery interceptors after buffer interceptors so the info logs in the same session
	// can be emitted for trouble shooting.
	builtInOpts = append(builtInOpts,
		grpc.ChainUnaryInterceptor(internal.RecoveryUnaryInterceptor(handler)),
		grpc.ChainStreamInterceptor(internal.RecoveryStreamInterceptor(handler)),
	)
	builtInOpts = append(builtInOpts, opts...)

	return grpc.NewServer(builtInOpts...)
}

// Run wraps start/stop of the gRPC server in a single run function
// with listening on multiple tcp and unix socket address.
//
// It also resister health and reflection services if the services have not registered.
func Run(server *grpc.Server, opts ...Option) func(context.Context) error { //nolint:cyclop,funlen,gocognit
	if server == nil {
		server = grpc.NewServer()
	}

	option := &options{}
	for _, opt := range opts {
		opt(option)
	}
	if len(option.addresses) == 0 {
		address := "localhost:8080"
		if a := os.Getenv("PORT"); a != "" {
			address = ":" + a
		}
		option.addresses = []socket{{network: "tcp", address: address}}
	}

	// Register config service if necessary.
	if option.configs != nil {
		pb.RegisterConfigServiceServer(server, internal.NewConfigServiceServer(option.configs))
	}
	// Register reflection service if necessary.
	if _, exist := server.GetServiceInfo()[grpc_reflection_v1.ServerReflection_ServiceDesc.ServiceName]; !exist {
		reflection.Register(server)
	}
	// Register health service if necessary.
	var healthServer *health.Server
	if _, exist := server.GetServiceInfo()[grpc_health_v1.Health_ServiceDesc.ServiceName]; !exist {
		healthServer = health.NewServer()
		defer healthServer.Resume()
		grpc_health_v1.RegisterHealthServer(server, healthServer)
	}

	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)

		defer context.AfterFunc(ctx, func() {
			slog.LogAttrs(ctx, slog.LevelInfo, "Starting shutdown gRPC Server...")
			if healthServer != nil {
				// Shutdown health server so client knows it's not serving.
				slog.LogAttrs(ctx, slog.LevelInfo, "Starting shutdown gRPC Health service...")
				healthServer.Shutdown()
				slog.LogAttrs(ctx, slog.LevelInfo, "Shutdown gRPC Health service completed.")
			}
			server.GracefulStop()
			slog.LogAttrs(ctx, slog.LevelInfo, "Shutdown gRPC Server completed.")
		})()

		slog.LogAttrs(ctx, slog.LevelInfo, "Starting gRPC Server...")
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(option.addresses))
		for _, addr := range option.addresses {
			addr := addr
			go func() {
				defer waitGroup.Done()

				if addr.network == "unix" {
					if err := os.RemoveAll(addr.address); err != nil {
						slog.LogAttrs(ctx, slog.LevelWarn, "Could not delete unix socket file.", slog.Any("error", err))
					}
				}
				listener, err := net.Listen(addr.network, addr.address)
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
		waitGroup.Wait()

		if err := context.Cause(ctx); err != nil && !errors.Is(err, ctx.Err()) {
			return err //nolint:wrapcheck
		}

		return nil
	}
}

func init() { //nolint:init
	// Redirect gRPC log to slog.
	grpclog.SetLoggerV2(internal.NewSlogger(slog.Default().Handler()))
}
