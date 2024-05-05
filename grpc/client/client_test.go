// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package client_test

import (
	"testing"

	ngrpc "github.com/nil-go/nilgo/grpc/client"
)

func TestAddGlobalDialOption(*testing.T) {
	// Should not panic.
	ngrpc.AddGlobalDialOption()
}
