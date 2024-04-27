// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package http_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/nil-go/sloth/sampling"

	nhttp "github.com/nil-go/nilgo/http"
	"github.com/nil-go/nilgo/http/internal/assert"
)

func TestBufferInterceptor(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		records     []slog.Record
		expected    string
	}{
		{
			description: "info only",
			records: []slog.Record{
				slog.NewRecord(time.Time{}, slog.LevelInfo, "info", 0),
			},
		},
		{
			description: "with error",
			records: []slog.Record{
				slog.NewRecord(time.Time{}, slog.LevelInfo, "info", 0),
				slog.NewRecord(time.Time{}, slog.LevelError, "error", 0),
			},
			expected: `level=INFO msg=info
level=ERROR msg=error
`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			logHandler := sampling.New(
				slog.NewTextHandler(buf, &slog.HandlerOptions{}),
				func(context.Context) bool { return false },
			)

			handler := http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
				for _, record := range testcase.records {
					_ = logHandler.Handle(request.Context(), record)
				}
			})

			writer := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/", nil)
			nhttp.BufferInterceptor(handler).ServeHTTP(writer, request)

			pwd, _ := os.Getwd()
			assert.Equal(t, testcase.expected, strings.ReplaceAll(buf.String(), pwd, ""))
		})
	}
}
