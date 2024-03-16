// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/nil-go/nilgo/internal/assert"
	"github.com/nil-go/nilgo/run"
)

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		runner      run.Runner
		ran         bool
		err         string
	}{
		{
			description: "empty runner",
			ran:         true,
		},
		{
			description: "run error",
			ran:         true,
			err:         "run error",
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
			ran:         true,
		},
		{
			description: "with start gate",
			runner:      run.New(run.WithStartGate(func(context.Context) error { return nil })),
			ran:         true,
		},
		{
			description: "start gate error",
			runner:      run.New(run.WithStartGate(func(context.Context) error { return errors.New("start gate error") })),
			err:         "start gate error",
		},
		{
			description: "with stop gate",
			runner:      run.New(run.WithStopGate(func(context.Context) error { return nil })),
			ran:         true,
		},
		{
			description: "stop gate error",
			runner:      run.New(run.WithStopGate(func(context.Context) error { return errors.New("stop gate error") })),
			ran:         true,
			err:         "stop gate error",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		var ran bool
		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

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

func TestRunner_Run_signal(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var runner run.Runner
	assert.NoError(t, runner.Run(ctx,
		func(ctx context.Context) error {
			timer := time.NewTimer(time.Minute)
			defer timer.Stop()
			select {
			case <-ctx.Done():
				return nil
			case <-timer.C:
				return errors.New("timeout")
			}
		},
		func(context.Context) error {
			return syscall.Kill(os.Getpid(), syscall.SIGINT)
		},
	))
}
