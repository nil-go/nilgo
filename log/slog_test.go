// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"

	"github.com/nil-go/sloth/sampling"

	"github.com/nil-go/nilgo/internal/assert"
	"github.com/nil-go/nilgo/log"
)

func TestNew(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		opts        []log.Option
		fn          func(context.Context, *slog.Logger)
		expected    string
	}{
		{
			description: "with handler",
			fn: func(ctx context.Context, logger *slog.Logger) {
				logger.InfoContext(ctx, "info log")
				logger.ErrorContext(ctx, "error log")
			},
			expected: `{"level":"INFO","msg":"info log"}
{"level":"ERROR","msg":"error log"}
`,
		},
		{
			description: "with sampler (info only)",
			opts: []log.Option{
				log.WithSampler(func(context.Context) bool { return false }),
			},
			fn: func(ctx context.Context, logger *slog.Logger) {
				logger.InfoContext(ctx, "info log")
			},
		},
		{
			description: "with sampler",
			opts: []log.Option{
				log.WithSampler(func(context.Context) bool { return false }),
			},
			fn: func(ctx context.Context, logger *slog.Logger) {
				logger.InfoContext(ctx, "info log")
				logger.ErrorContext(ctx, "error log")
			},
			expected: `{"level":"INFO","msg":"info log"}
{"level":"ERROR","msg":"error log"}
`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			var buf bytes.Buffer
			logger := log.New(append(testcase.opts, log.WithHandler(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
				ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
					if a.Key == slog.TimeKey && len(groups) == 0 {
						return slog.Attr{}
					}

					return a
				},
			})))...)

			ctx, cancel := sampling.WithBuffer(context.Background())
			defer cancel()
			testcase.fn(ctx, logger)

			assert.Equal(t, testcase.expected, buf.String())
		})
	}
}
