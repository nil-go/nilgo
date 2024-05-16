// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package nilgo_test

import (
	"context"
	"errors"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/nil-go/nilgo"
	"github.com/nil-go/nilgo/internal/assert"
)

func TestRunner_Run(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		runner      nilgo.Runner
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
			runner: nilgo.New(nilgo.WithPreRun(func(ctx context.Context) error {
				<-ctx.Done()

				return nil
			})),
			ran: true,
		},
		{
			description: "pre-run error",
			runner:      nilgo.New(nilgo.WithPreRun(func(context.Context) error { return errors.New("pre-run error") })),
			err:         "pre-run error",
			ran:         true,
		},
		{
			description: "with post-run",
			runner: nilgo.New(nilgo.WithPostRun(func(context.Context) error {
				return nil
			})),
			ran: true,
		},
		{
			description: "post-run error",
			runner: nilgo.New(nilgo.WithPostRun(func(context.Context) error {
				return errors.New("post-run error")
			}),
			),
			err: "post-run error",
			ran: true,
		},
		{
			description: "with start gate",
			runner:      nilgo.New(nilgo.WithStartGate(func(context.Context) error { return nil })),
			ran:         true,
		},
		{
			description: "start gate error",
			runner:      nilgo.New(nilgo.WithStartGate(func(context.Context) error { return errors.New("start gate error") })),
			err:         "start gate error",
		},
		{
			description: "with stop gate",
			runner:      nilgo.New(nilgo.WithStopGate(func(context.Context) error { return nil })),
			ran:         true,
		},
		{
			description: "stop gate error",
			runner:      nilgo.New(nilgo.WithStopGate(func(context.Context) error { return errors.New("stop gate error") })),
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

	startTime := time.Now()
	var ran bool
	runner := nilgo.New(
		nilgo.WithPostRun(func(context.Context) error {
			ran = ctx.Err() == nil

			return nil
		}),
	)
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
	assert.Equal(t, true, ran)
	assert.Equal(t, true, time.Since(startTime) < time.Minute)
}

func TestRunner_Run_cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startTime := time.Now()
	var ran bool
	runner := nilgo.New(
		nilgo.WithPostRun(func(ctx context.Context) error {
			ran = ctx.Err() == nil

			return nil
		}),
	)
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
			cancel()

			return nil
		},
	))
	assert.Equal(t, true, ran)
	assert.Equal(t, true, time.Since(startTime) < time.Minute)
}
