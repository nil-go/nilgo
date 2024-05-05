// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package nilgo

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime"
	"time"
)

// PProf starts a pprof server at localhost:6060.
//
// If port 6060 is not available, it will try to find an available port.
func PProf(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)
	server := &http.Server{
		Handler:     mux,
		ReadTimeout: time.Second,
	}

	defer context.AfterFunc(ctx, func() {
		slog.LogAttrs(ctx, slog.LevelInfo, "Starting shutdown pprof Server...")
		if err := server.Shutdown(context.WithoutCancel(ctx)); err != nil {
			slog.LogAttrs(ctx, slog.LevelWarn, "Fail to shutdown pprof server.", slog.Any("error", err))
		}
		slog.LogAttrs(ctx, slog.LevelInfo, "Shutdown pprof Server completed.")
	})()

	slog.LogAttrs(ctx, slog.LevelInfo, "Starting pprof server.")
	listener, err := net.Listen("tcp", "localhost:6060")
	if err != nil {
		listener, err = net.Listen("tcp", "localhost:0")
		if err != nil {
			slog.LogAttrs(ctx, slog.LevelWarn, "Fail to find port for pprof server.", slog.Any("error", err))

			return nil
		}
	}
	slog.LogAttrs(ctx, slog.LevelInfo, fmt.Sprintf("pprof server started at http://%s/debug/pprof/.", listener.Addr()))

	runtime.SetBlockProfileRate(1) // Required by gRPC server.
	if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.LogAttrs(ctx, slog.LevelWarn, "Fail to start pprof server.", slog.Any("error", err))
	}

	return nil
}
