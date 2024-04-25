// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package nilgo_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"testing"
	"testing/fstest"

	"github.com/nil-go/konf"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/config"
	"github.com/nil-go/nilgo/internal/assert"
	"github.com/nil-go/nilgo/log"
	"github.com/nil-go/nilgo/run"
)

func TestRun(t *testing.T) {
	var (
		buf     bytes.Buffer
		started bool
	)

	err := nilgo.Run(
		slog.NewJSONHandler(&buf, &slog.HandlerOptions{
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key == slog.TimeKey && len(groups) == 0 {
					return slog.Attr{}
				}

				return a
			},
		}),
		log.WithSampler(func(context.Context) bool { return true }),
		config.WithFS(fstest.MapFS{"config/config.yaml": {Data: []byte("nilgo:\n  source: fs")}}),
		run.WithPreRun(func(context.Context) error {
			started = true

			return nil
		}),
		func(context.Context) error {
			slog.Info("info log", "source", konf.Get[string]("nilgo.source"))

			return nil
		},
	)

	assert.NoError(t, err)
	assert.Equal(t, true, started)
	assert.Equal(t, `{"level":"INFO","msg":"info log","source":"fs"}
`, buf.String())
}

func TestRun_error(t *testing.T) {
	testcases := []struct {
		description string
		args        []any
		err         string
	}{
		{
			description: "unknown argument type",
			args:        []any{"unknown"},
			err:         "unknown argument type: string",
		},
		{
			description: "config error",
			args: []any{
				config.With(func(*konf.Config) error {
					return errors.New("config error")
				}),
			},
			err: "init config: config error",
		},
		{
			description: "runner error",
			args: []any{
				func(context.Context) error {
					return errors.New("runner error")
				},
			},
			err: "run: runner error",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			err := nilgo.Run(testcase.args...)
			assert.Equal(t, testcase.err, err.Error())
		})
	}
}
