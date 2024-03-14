// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log

import (
	"log/slog"

	"google.golang.org/grpc"
	"google.golang.org/grpc/grpclog"
)

// ServerOptions returns gRPC server options for log related interceptors,
// which includes panic recovery, buffer for sampling.Handler.
// It also replaces the default gRPC logger with the provided slog.Handler.
func ServerOptions(handler slog.Handler) []grpc.ServerOption {
	if handler == nil {
		handler = slog.Default().Handler()
	}
	grpclog.SetLoggerV2(NewSlogger(handler))

	var (
		unaryInterceptors  []grpc.UnaryServerInterceptor
		streamInterceptors []grpc.StreamServerInterceptor
	)

	if isSamplingHandler(handler) {
		unaryInterceptors = append(unaryInterceptors, BufferUnaryInterceptor)
		streamInterceptors = append(streamInterceptors, BufferStreamInterceptor)
	}

	// Add recovery interceptors after buffer interceptors so the info logs in the same session
	// can be emitted for trouble shooting.
	unaryInterceptors = append(unaryInterceptors, RecoveryUnaryInterceptor(handler))
	streamInterceptors = append(streamInterceptors, RecoveryStreamInterceptor(handler))

	return []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptors...),
		grpc.ChainStreamInterceptor(streamInterceptors...),
	}
}
