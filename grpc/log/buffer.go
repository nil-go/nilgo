// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log

import (
	"context"
	"log/slog"
	"reflect"
	"unsafe"

	"github.com/nil-go/sloth/sampling"
	"google.golang.org/grpc"
)

// BufferUnaryInterceptor returns a unary interceptor that inserts log buffer
// into context.Context for sampling.Handler.
func BufferUnaryInterceptor(
	ctx context.Context,
	req any,
	_ *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	ctx, put := sampling.WithBuffer(ctx)
	defer put()

	return handler(ctx, req)
}

// BufferStreamInterceptor returns a stream interceptor that inserts log buffer
// into context.Context for sampling.Handler.
func BufferStreamInterceptor(
	srv any,
	stream grpc.ServerStream,
	_ *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) error {
	ctx, put := sampling.WithBuffer(stream.Context())
	defer put()

	return handler(srv, serverStream{ServerStream: stream, ctx: ctx})
}

type serverStream struct {
	grpc.ServerStream

	ctx context.Context //nolint:containedctx
}

func (s serverStream) Context() context.Context { return s.ctx }

func isSamplingHandler(handler slog.Handler) bool {
	switch handler.(type) {
	case sampling.Handler, *sampling.Handler:
		return true
	default:
	}

	value := reflect.ValueOf(handler)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return false
	}

	// Check nested/embedded `slog.Handler`.
	valueCopy := reflect.New(value.Type()).Elem()
	valueCopy.Set(value)
	for _, name := range []string{"handler", "Handler"} {
		if v := valueCopy.FieldByName(name); v.IsValid() {
			v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
			if h, ok := v.Interface().(slog.Handler); ok {
				return isSamplingHandler(h)
			}
		}
	}

	return false
}
