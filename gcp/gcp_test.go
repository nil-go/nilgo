// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package gcp_test

import (
	"context"
	"log/slog"
	"testing"

	sgcp "github.com/nil-go/sloth/gcp"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/nil-go/nilgo/gcp"
	"github.com/nil-go/nilgo/gcp/profiler"
)

func TestOptions(t *testing.T) {
	t.Parallel()

	testcase := []struct {
		description string
		opts        []gcp.Option
		assertion   func(*testing.T, []any)
	}{
		{
			description: "with project",
			opts: []gcp.Option{
				gcp.WithProject("project"),
				gcp.WithService("service"),
				gcp.WithVersion("version"),
			},
			assertion: func(t *testing.T, opts []any) {
				t.Helper()

				assert.Len(t, opts, 1)
				_, ok := opts[0].(slog.Handler)
				assert.True(t, ok)
			},
		},
		{
			description: "with options",
			opts: []gcp.Option{
				gcp.WithProject("project"),
			},
			assertion: func(t *testing.T, opts []any) {
				t.Helper()

				assert.Len(t, opts, 1)
				_, ok := opts[0].(slog.Handler)
				assert.True(t, ok)
			},
		},
		{
			description: "with profiler",
			opts: []gcp.Option{
				gcp.WithProject("project"),
				gcp.WithProfiler(profiler.WithMutexProfiling()),
			},
			assertion: func(t *testing.T, opts []any) {
				t.Helper()

				assert.Len(t, opts, 2)
				_, ok := opts[0].(slog.Handler)
				assert.True(t, ok)
				_, ok = opts[1].(func(context.Context) error)
				assert.True(t, ok)
			},
		},
		{
			description: "with trace",
			opts: []gcp.Option{
				gcp.WithProject("project"),
				gcp.WithTrace(),
			},
			assertion: func(t *testing.T, opts []any) {
				t.Helper()

				assert.Len(t, opts, 2)
				_, ok := opts[0].(slog.Handler)
				assert.True(t, ok)
				_, ok = opts[1].(*trace.TracerProvider)
				assert.True(t, ok)
			},
		},
		{
			description: "with metric",
			opts: []gcp.Option{
				gcp.WithProject("project"),
				gcp.WithMetric(),
			},
			assertion: func(t *testing.T, opts []any) {
				t.Helper()

				assert.Len(t, opts, 2)
				_, ok := opts[0].(slog.Handler)
				assert.True(t, ok)
				_, ok = opts[1].(*metric.MeterProvider)
				assert.True(t, ok)
			},
		},
		{
			description: "without project",
			opts: []gcp.Option{
				gcp.WithLog(sgcp.WithLevel(slog.LevelError)),
				gcp.WithProfiler(),
			},
			assertion: func(t *testing.T, opts []any) {
				t.Helper()

				assert.Len(t, opts, 1)
				_, ok := opts[0].(slog.Handler)
				assert.True(t, ok)
			},
		},
	}

	for _, testcase := range testcase {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			opts, err := gcp.Options(testcase.opts...)
			require.NoError(t, err)
			testcase.assertion(t, opts)
		})
	}
}
