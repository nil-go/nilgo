// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package gcp_test

import (
	"context"
	"log/slog"
	"testing"

	sgcp "github.com/nil-go/sloth/gcp"
	"github.com/stretchr/testify/assert"

	"github.com/nil-go/nilgo/gcp"
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
				gcp.WithOptions(func() []gcp.Option {
					return []gcp.Option{
						gcp.WithProject("project"),
					}
				}),
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
				gcp.WithProfiler(),
				gcp.WithMutextProfiling(),
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
			description: "without project",
			opts: []gcp.Option{
				gcp.WithLogOptions(sgcp.WithLevel(slog.LevelError)),
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

			testcase.assertion(t, gcp.Options(testcase.opts...))
		})
	}
}
