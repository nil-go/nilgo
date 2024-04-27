// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log

import (
	"context"
	"fmt"
	"log/slog"
	"runtime"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryInterceptor returns a unary interceptor that recovers from gRPC handler panic.
func RecoveryUnaryInterceptor(logHandler slog.Handler) func(
	context.Context, any, *grpc.UnaryServerInfo, grpc.UnaryHandler,
) (any, error) {
	return func( //nolint:nonamedreturns
		ctx context.Context,
		req any,
		_ *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (_ any, err error) {
		defer func() {
			if r := recover(); r != nil {
				err = logPanic(ctx, logHandler, r)
			}
		}()

		return handler(ctx, req)
	}
}

// RecoveryStreamInterceptor returns a stream interceptor that recovers from gRPC handler panic.
func RecoveryStreamInterceptor(logHandler slog.Handler) func(
	any, grpc.ServerStream, *grpc.StreamServerInfo, grpc.StreamHandler,
) error {
	return func( //nolint:nonamedreturns
		srv any,
		stream grpc.ServerStream,
		_ *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = logPanic(stream.Context(), logHandler, r)
			}
		}()

		return handler(srv, stream)
	}
}

func logPanic(ctx context.Context, handler slog.Handler, message any) error {
	err, ok := message.(error)
	if !ok {
		err = fmt.Errorf("%v", message) //nolint:goerr113
	}

	var pcs [1]uintptr
	runtime.Callers(4, pcs[:]) //nolint:gomnd // Skip runtime.Callers, panic, this function and interceptor.
	r := slog.NewRecord(time.Now(), slog.LevelError, "Panic Recovered", pcs[0])
	r.AddAttrs(slog.Any("error", err))
	_ = handler.Handle(ctx, r) // Ignore error: It's fine to lose log.

	return status.Error(codes.Internal, "") //nolint:wrapcheck // Truncate message for security concerns.
}
