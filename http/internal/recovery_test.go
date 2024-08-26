// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/nil-go/nilgo/http/internal"
	"github.com/nil-go/nilgo/http/internal/assert"
)

func TestRecoveryInterceptor(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		handler     func(http.ResponseWriter, *http.Request)
		code        int
		log         string
	}{
		{
			description: "normal",
			handler:     func(http.ResponseWriter, *http.Request) {},
			code:        http.StatusOK,
		},
		{
			description: "panic",
			handler:     func(http.ResponseWriter, *http.Request) { panic("panic from handler") },
			code:        http.StatusInternalServerError,
			log: `level=ERROR source=/recovery_test.go:35 msg="Panic Recovered" error="panic from handler"
`,
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
				AddSource: true,
				ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
					if len(groups) == 0 && attr.Key == slog.TimeKey {
						return slog.Attr{}
					}

					return attr
				},
			})

			writer := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			internal.RecoveryInterceptor(http.HandlerFunc(testcase.handler), handler).ServeHTTP(writer, req)

			assert.Equal(t, testcase.code, writer.Code)
			pwd, _ := os.Getwd()
			assert.Equal(t, testcase.log, strings.ReplaceAll(buf.String(), pwd, ""))
		})
	}
}
