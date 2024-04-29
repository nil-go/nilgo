// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal_test

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

	"github.com/nil-go/nilgo/http/internal"
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
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			internal.BufferInterceptor(handler).ServeHTTP(writer, req)

			pwd, _ := os.Getwd()
			assert.Equal(t, testcase.expected, strings.ReplaceAll(buf.String(), pwd, ""))
		})
	}
}

func TestIsSamplingHandler(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		handler     slog.Handler
		expected    bool
	}{
		{
			description: "nil handler",
		},
		{
			description: "sampling handler",
			handler:     sampling.Handler{},
			expected:    true,
		},
		{
			description: "sampling handler (pointer)",
			handler:     &sampling.Handler{},
			expected:    true,
		},
		{
			description: "embed sampling handler",
			handler:     handlerEmbed{sampling.Handler{}},
			expected:    true,
		},
		{
			description: "embed sampling handler (pointer)",
			handler:     &handlerEmbed{sampling.Handler{}},
			expected:    true,
		},
		{
			description: "nested sampling handler",
			handler:     handlerWrapper{handler: sampling.Handler{}},
			expected:    true,
		},
		{
			description: "nested sampling handler (pointer)",
			handler:     &handlerWrapper{handler: sampling.Handler{}},
			expected:    true,
		},
		{
			description: "deep nested sampling handler",
			handler:     handlerWrapper{handler: handlerWrapper{handler: sampling.Handler{}}},
			expected:    true,
		},
		{
			description: "non sampling handler",
			handler:     slogDiscard{},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testcase.expected, internal.IsSamplingHandler(testcase.handler))
		})
	}
}

type handlerWrapper struct {
	slogDiscard
	handler slog.Handler
}

type slogDiscard struct{}

func (slogDiscard) Enabled(context.Context, slog.Level) bool  { return false }
func (slogDiscard) Handle(context.Context, slog.Record) error { return nil }
func (slogDiscard) WithAttrs([]slog.Attr) slog.Handler        { return slogDiscard{} }
func (slogDiscard) WithGroup(string) slog.Handler             { return slogDiscard{} }

type handlerEmbed struct {
	slog.Handler
}
