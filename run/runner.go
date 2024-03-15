// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run

import (
	"context"
	"errors"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
)

// Runner is a pre-configured runtime for executing runs in parallel.
//
// To create an Runner, use [New].
type Runner struct {
	preRuns    []func(context.Context) error
	startGates []func(context.Context) error
	running    atomic.Bool
}

// New creates a new Runner with the given Option(s).
func New(opts ...Option) *Runner {
	option := &options{}
	for _, opt := range opts {
		opt(option)
	}

	return (*Runner)(option)
}

// Run executes the given run in parallel with well-configured runtime,
// which includes logging, configuration, and telemetry.
//
// The run running in parallel without any explicit order,
// which means it should not have temporal dependencies between each other.
//
// The execution can be interrupted if any run returns non-nil error,
// or it receives an OS signal syscall.SIGINT or syscall.SIGTERM.
// It waits all run return unless it's forcefully terminated by OS.
func (e *Runner) Run(ctx context.Context, runs ...func(context.Context) error) error {
	if e == nil {
		// Use empty instance instead to avoid nil pointer dereference,
		// Assignment propagates only to callee but not to caller.
		e = &Runner{}
	}

	// Prevent the runner from running concurrently.
	if e.running.Swap(true) {
		return errors.New("runner is already running") //nolint:goerr113
	}
	defer e.running.Store(false)

	allRuns := make([]func(context.Context) error, 0, len(e.preRuns)+1)
	// Add gate to wait for all pre-runs to start.
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(e.preRuns))
	for _, run := range e.preRuns {
		run := run
		allRuns = append(allRuns,
			func(ctx context.Context) error {
				defer waitGroup.Done()

				return run(ctx)
			},
		)
	}
	e.startGates = append(e.startGates,
		func(context.Context) error {
			waitGroup.Wait()

			return nil
		},
	)

	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	allRuns = append(allRuns,
		func(ctx context.Context) (err error) { //nolint:nonamedreturns
			defer func() {
				cancel(err)
			}()

			// Terminate signals apply to the runs, then cancel the root context for pre-runs.
			nctx, ncancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
			defer ncancel()

			// Wait for all startGates to open.
			if err = Parallel(nctx, e.startGates...); err != nil {
				return err
			}

			return Parallel(nctx, runs...)
		},
	)

	return Parallel(ctx, allRuns...)
}
