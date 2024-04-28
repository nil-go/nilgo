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
	var opts []any
	if metadata.OnGCE() {
		opts = gcp.Options()
	}
	opts = append(opts,
		config.WithFS(configFS),
		nhttp.Run(&http.Server{ReadTimeout: time.Second}),
	)

	if err := nilgo.Run(opts...); err != nil {
		panic(err)
	}
}
