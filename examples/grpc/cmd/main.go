// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"embed"

	"cloud.google.com/go/compute/metadata"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/config"
	"github.com/nil-go/nilgo/gcp"
	ngrpc "github.com/nil-go/nilgo/grpc"
)

//go:embed config
var configFS embed.FS

func main() {
	var opts []any
	switch {
	case metadata.OnGCE():
		opts = gcp.Options()
	default:
		opts = []any{nilgo.PProf}
	}
	opts = append(opts,
		config.WithFS(configFS),
		ngrpc.Run(ngrpc.NewServer()),
	)

	if err := nilgo.Run(opts...); err != nil {
		panic(err)
	}
}
