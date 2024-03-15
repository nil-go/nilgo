// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run

import "context"

// WithPreRun provides runs to execute before the main runs provided in Runner.Run.
//
// It's guaranteed that all goroutines for pre-runs start before the main runs start,
// and end after the main runs end if it's blocking with [context.Context.Done].
func WithPreRun(runs ...func(context.Context) error) Option {
	return func(opts *options) {
		opts.preRuns = append(opts.preRuns, runs...)
	}
}

// WithStartGate provides gates to block the start of main runs provided in Runner.Run,
// until all start gates returns without error.
//
// All start gates must return in limited time to avoid blocking the main runs.
func WithStartGate(gates ...func(context.Context) error) Option {
	return func(opts *options) {
		opts.startGates = append(opts.startGates, gates...)
	}
}

type (
	// Option configures the Runner with specific options.
	Option  func(*options)
	options Runner
)
