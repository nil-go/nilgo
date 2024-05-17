// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package main

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"cloud.google.com/go/compute/metadata"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/dev"
	"github.com/nil-go/nilgo/gcp/log"
	"github.com/nil-go/nilgo/gcp/profiler"
	nhttp "github.com/nil-go/nilgo/http"
	"github.com/nil-go/nilgo/otlp"
)

func main() {
	var (
		opts []nilgo.Option
		runs []func(context.Context) error
	)
	switch {
	case metadata.OnGCE():
		opts = append(opts, nilgo.WithLogHandler(log.Handler()))
		traceProvider, err := otlp.TraceProvider()
		if err != nil {
			panic(err)
		}
		opts = append(opts, nilgo.WithTraceProvider(traceProvider))
		meterProvider, err := otlp.MeterProvider()
		if err != nil {
			panic(err)
		}
		opts = append(opts, nilgo.WithMeterProvider(meterProvider))
		runs = append(runs, profiler.Run(profiler.WithMutexProfiling()))
	default:
		runs = append(runs, dev.Pprof)
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
	runs = append(runs,
		nhttp.Run(server, nhttp.WithConfigService()),
	)

	if err := nilgo.New(opts...).Run(context.Background(), runs...); err != nil {
		panic(err)
	}
}
