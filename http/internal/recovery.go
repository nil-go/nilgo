// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"runtime"
	"time"
)

// RecoveryInterceptor returns a server interceptor that recovers from HTTP handler panic.
// It overrides default recovery in [http] to provide better log integration.
func RecoveryInterceptor(handler http.Handler, logHandler slog.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		defer func(ctx context.Context) {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r) //nolint:err113
				}

				var pcs [1]uintptr
				runtime.Callers(3, pcs[:]) //nolint:mnd // Skip runtime.Callers, panic, this function.
				record := slog.NewRecord(time.Now(), slog.LevelError, "Panic Recovered", pcs[0])
				record.AddAttrs(slog.Any("error", err))
				_ = logHandler.Handle(ctx, record) // Ignore error: It's fine to lose log.

				http.Error(writer, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
		}(request.Context())

		handler.ServeHTTP(writer, request)
	})
}
