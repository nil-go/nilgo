// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"embed"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/config"
	"github.com/nil-go/nilgo/gcp"
	nhttp "github.com/nil-go/nilgo/http"
)

//go:embed config
var configFS embed.FS

func main() {
	var args []any
	switch {
	case metadata.OnGCE():
		opts, err := gcp.Options(
			gcp.WithTrace(),
			gcp.WithMetric(),
			gcp.WithProfiler(),
		)
		if err != nil {
			panic(err)
		}
		args = append(args, opts...)
	default:
		args = append(args, nilgo.PProf)
	}
	args = append(args,
		config.WithFS(configFS),
		nhttp.Run(&http.Server{ReadTimeout: time.Second}),
	)

	if err := nilgo.Run(args...); err != nil {
		panic(err)
	}
}
