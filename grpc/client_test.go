// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package grpc_test

import (
	"testing"

	ngrpc "github.com/nil-go/nilgo/grpc"
)

func TestLinkName(*testing.T) {
	// Should not panic.
	ngrpc.WithClientTelemetry()
}
