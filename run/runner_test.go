// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nil-go/nilgo/internal/assert"
	"github.com/nil-go/nilgo/run"
)

func TestRunner_Run(t *testing.T) {
	for _, testcase := range testcases() {
		testcase := testcase

		var ran bool
		t.Run(testcase.description, func(t *testing.T) {
			err := testcase.runner.Run(
				context.Background(),
				func(context.Context) error {
					ran = true
					if testcase.err != "" {
						return errors.New(testcase.err)
					}

					return nil
				},
			)

			assert.Equal(t, testcase.ran, ran)
			if testcase.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testcase.err)
			}
		})
	}
}

func testcases() []struct {
	description string
	runner      *run.Runner
	ran         bool
	err         string
} {
	return []struct {
		description string
		runner      *run.Runner
		ran         bool
		err         string
	}{
		{
			description: "nil runner",
			ran:         true,
		},
		{
			description: "run error",
			ran:         true,
			err:         "run error",
		},
		{
			description: "with gate",
			runner:      run.New(run.WithGate(func(context.Context) error { return nil })),
			ran:         true,
		},
		{
			description: "gate error",
			runner:      run.New(run.WithGate(func(context.Context) error { return errors.New("gate error") })),
			err:         "gate error",
		},
		{
			description: "with pre-run",
			runner:      run.New(run.WithPreRun(func(context.Context) error { return nil })),
			ran:         true,
		},
		{
			description: "pre-run error",
			runner:      run.New(run.WithPreRun(func(context.Context) error { return errors.New("pre-run error") })),
			err:         "pre-run error",
		},
	}
}

func TestRunner_Run_twice(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runner run.Runner
	// As there is no run, it should return immediately.
	assert.NoError(t, runner.Run(ctx))

	started := make(chan struct{})
	go func() {
		err := runner.Run(ctx, func(context.Context) error {
			close(started)
			<-ctx.Done()

			return nil
		})
		assert.NoError(t, err)
	}()
	<-started

	// It should return an error as it's already running.
	assert.EqualError(t, runner.Run(ctx), "runner is already running")
}
