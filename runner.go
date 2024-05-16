// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package nilgo provides a simple way to bootstrap an production-ready application.
package nilgo

import (
	"context"
	"errors"
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
	postRuns   []func(context.Context) error
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
//
// The execution flow is as follows:
// 1. Starts all pre runs and start gates in parallel.
// 2. Waits for all pre runs start and start gates complete.
// 3. Starts all main runs.
// 4. Waits for OS interrupt or terminal signals or all main runs complete.
// 5. Waits for all stop gates complete.
// 6. Stop all main runs.
// 7. Waits for all post runs complete.
func (r Runner) Run(ctx context.Context, runs ...func(context.Context) error) error { //nolint:funlen
	preRuns := make([]func(context.Context) error, 0, len(r.preRuns))
	startGates := slices.Clone(r.startGates)
	if len(r.preRuns) > 0 {
		// Append wait group to wait for all pre runs to start.
		var waitGroup sync.WaitGroup
		waitGroup.Add(len(r.preRuns))
		for _, run := range r.preRuns {
			run := run
			preRuns = append(preRuns,
				func(ctx context.Context) error {
					waitGroup.Done()

					return run(ctx)
				},
			)
		}
		// Add gate to wait for all pre runs to start.
		startGates = append(startGates,
			func(context.Context) error {
				waitGroup.Wait()

				return nil
			},
		)
	}

	// Context can be terminated by either OS signals or cancellation on ctx.
	signalCtx, signalCancel := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	defer signalCancel()

	// Root context which is used for pre/post runs.
	// It does not propagate the cancellation from ctx.
	// It depends on signalCtx for cancellation.
	rootCtx, rootCancel := context.WithCancelCause(context.WithoutCancel(ctx))
	defer rootCancel(nil)
	// Context is used for main runs with start/stop gates.
	runCtx, runCancel := context.WithCancel(rootCtx)
	defer runCancel()

	return parallel(rootCtx,
		append(preRuns,
			func(context.Context) error {
				defer runCancel() // Notify all main runs to stop.

				<-signalCtx.Done()

				// Wait for all stop gates to open.
				return parallel(runCtx, r.stopGates...)
			},
			func(context.Context) (err error) { //nolint:nonamedreturns
				defer func() {
					signalCancel() // Stop listening to OS signals.

					// Wait for all post runs to finish.
					e := parallel(rootCtx, r.postRuns...)
					if err == nil {
						err = e
					}
					rootCancel(err) // Notify all pre runs to stop.
				}()

				// Wait for all start gates to open.
				if err = parallel(runCtx, startGates...); err != nil {
					return err
				}

				return parallel(runCtx, runs...)
			},
		)...,
	)
}

func parallel(ctx context.Context, runs ...func(context.Context) error) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(runs))
	for _, run := range runs {
		run := run
		go func() {
			defer waitGroup.Done()

			if err := run(ctx); err != nil {
				cancel(err)
			}
		}()
	}
	waitGroup.Wait()

	if err := context.Cause(ctx); err != nil && !errors.Is(err, ctx.Err()) {
		return err //nolint:wrapcheck
	}

	return nil
}
