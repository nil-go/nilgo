// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run

import "context"

// WithPreRun provides runs to execute before the main runs provided in Runner.Run.
//
// It's guaranteed that all goroutines for pre-runs start before the main runs start,
// and end after the main runs end if it's blocking with [context.Context.Done].
func WithPreRun(runs ...func(context.Context) error) Option {
	return func(options *options) {
		options.preRuns = append(options.preRuns, runs...)
	}
}

// WithPostRun provides runs to execute after the main runs provided in Runner.Run.
func WithPostRun(runs ...func(context.Context) error) Option {
	return func(options *options) {
		options.postRuns = append(options.postRuns, runs...)
	}
}

// WithStartGate provides gates to block the start of main runs provided in Runner.Run,
// until all start gates returns without error.
//
// All start gates must return in limited time to avoid blocking the main runs.
func WithStartGate(gates ...func(context.Context) error) Option {
	return func(options *options) {
		options.startGates = append(options.startGates, gates...)
	}
}

// WithStopGate provides gates to block the stop of main runs provided in Runner.Run,
// until all stop gates returns.
//
// All stop gates must return in limited time to avoid blocking the main runs.
func WithStopGate(gates ...func(context.Context) error) Option {
	return func(options *options) {
		options.stopGates = append(options.stopGates, gates...)
	}
}

type (
	// Option configures the Runner with specific options.
	Option  func(*options)
	options Runner
)
