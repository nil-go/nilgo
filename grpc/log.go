// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package grpc

import (
	"log/slog"

	"google.golang.org/grpc/grpclog"

	"github.com/nil-go/nilgo/grpc/internal"
)

func init() { //nolint:init
	// Redirect gRPC log to slog.
	grpclog.SetLoggerV2(internal.NewSlogger(slog.Default().Handler()))
}
