// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal

import (
	"log/slog"
	"net/http"
	"reflect"
	"unsafe"

	"github.com/nil-go/sloth/sampling"
)

// BufferInterceptor returns a http interceptor that inserts log buffer
// into context.Context for sampling.Handler.
func BufferInterceptor(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx, put := sampling.WithBuffer(request.Context())
		defer put()

		handler.ServeHTTP(writer, request.WithContext(ctx))
	})
}

func IsSamplingHandler(handler slog.Handler) bool {
	switch handler.(type) {
	case sampling.Handler, *sampling.Handler:
		return true
	default:
	}

	value := reflect.ValueOf(handler)
	if value.Kind() == reflect.Pointer {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return false
	}

	// Check nested/embedded `slog.Handler`.
	valueCopy := reflect.New(value.Type()).Elem()
	valueCopy.Set(value)
	for _, name := range []string{"handler", "Handler"} {
		if v := valueCopy.FieldByName(name); v.IsValid() {
			v = reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
			if h, ok := v.Interface().(slog.Handler); ok {
				return IsSamplingHandler(h)
			}
		}
	}

	return false
}
