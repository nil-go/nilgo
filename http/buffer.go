// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package http

import (
	"net/http"

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
