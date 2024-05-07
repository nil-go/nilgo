// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"embed"
	"log/slog"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/config"
	"github.com/nil-go/nilgo/dev"
	"github.com/nil-go/nilgo/gcp"
	nhttp "github.com/nil-go/nilgo/http"
)

//go:embed config
var configFS embed.FS

func main() {
	var args []any
	switch {
	case metadata.OnGCE():
		opts, err := gcp.Args(
			gcp.WithLog(),
			gcp.WithTrace(),
			gcp.WithMetric(),
			gcp.WithProfiler(),
		)
		if err != nil {
			panic(err)
		}
		args = append(args, opts...)
	default:
		args = append(args, dev.Pprof)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		slog.InfoContext(r.Context(), "hello world")
		if _, err := w.Write([]byte("Hello, World!")); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	server := &http.Server{
		Handler: otelhttp.NewHandler(mux, "", otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			if operation != "" {
				return operation
			}

			method := r.Method
			if method == "" {
				method = "GET"
			}

			return method + " " + r.URL.Path
		})),
		ReadTimeout: time.Second,
	}
	args = append(args,
		config.WithFS(configFS),
		nhttp.Run(
			server,
			nhttp.WithConfigService(),
		),
	)

	if err := nilgo.Run(args...); err != nil {
		panic(err)
	}
}
