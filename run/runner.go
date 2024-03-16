// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run

import (
	"context"
	"os/signal"
	"slices"
	"sync"
	"syscall"
)

// Runner is a pre-configured runtime for executing runs in parallel.
//
// To create an Runner, use [New].
type Runner struct {
	preRuns    []func(context.Context) error
	startGates []func(context.Context) error
	stopGates  []func(context.Context) error
}

// New creates a new Runner with the given Option(s).
func New(opts ...Option) Runner {
	option := &options{}
	for _, opt := range opts {
		opt(option)
	}

	return Runner(*option)
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
func (e Runner) Run(ctx context.Context, runs ...func(context.Context) error) error {
	allRuns := make([]func(context.Context) error, 0, len(e.preRuns)+1)
	startGates := slices.Clone(e.startGates)
	if len(e.preRuns) > 0 {
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
		startGates = append(startGates,
			func(context.Context) error {
				waitGroup.Wait()

				return nil
			},
		)
	}

	// Root context which is used for pre-runs.
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)
	// Context can be terminated by OS signals, which is used for start-gates and parent of context for main runs.
	signalCtx, signalCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()
	// Context is used for main runs and stop-gates.
	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()

	allRuns = append(allRuns,
		func(context.Context) error {
			defer runCancel()

			<-signalCtx.Done()

			return Parallel(runCtx, e.stopGates...)
		},
		func(context.Context) (err error) { //nolint:nonamedreturns
			defer func() {
				cancel(err)
			}()

			// Wait for all startGates to open.
			// Use signalCtx to allow it to be interrupted by OS signals.
			if err = Parallel(signalCtx, startGates...); err != nil {
				return err
			}

			return Parallel(runCtx, runs...)
		},
	)

	return Parallel(ctx, allRuns...)
}
