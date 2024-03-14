// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package grpc

import (
	_ "unsafe" // For go:linkname

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

//go:linkname addGlobalDialOptions google.golang.org/grpc/internal.AddGlobalDialOptions
var addGlobalDialOptions any //nolint:gochecknoglobals // func(opt ...DialOption)

// WithDialOption adds global dial options for all gRPC clients.
//
// CAUTION: This function may break in new version of `google.golang.org/grpc`
// since it is using internal package from grpc.
func WithDialOption(opts ...grpc.DialOption) {
	addGlobalDialOptions.(func(opt ...grpc.DialOption))(opts...) //nolint:forcetypeassert
}

// WithClientTelemetry enables trace and metrics instruments on all gRPC clients.
//
// CAUTION: This function may break in new version of `google.golang.org/grpc`
// since it is using internal package from grpc.
func WithClientTelemetry(opts ...otelgrpc.Option) {
	WithDialOption(grpc.WithStatsHandler(otelgrpc.NewClientHandler(opts...)))
}
