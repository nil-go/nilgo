// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log

import (
	"context"
	"log/slog"
)

// WithHandler provides a customized slog.Handler.
//
// By default, it uses the default handler in [slog].
func WithHandler(handler slog.Handler) Option {
	return func(options *options) {
		options.handler = handler
	}
}

// WithSampler provides a sampler function which decides whether Info logs should write to output.
//
// By default, it disables simpling with nil sampler.
func WithSampler(sampler func(context.Context) bool) Option {
	return func(options *options) {
		options.sampler = sampler
	}
}

// WithLogAsTraceEvent enables logging as trace event.
//
// It could significantly reduce the log volume then cost as trace is priced by number of span.
func WithLogAsTraceEvent() Option {
	return func(options *options) {
		options.asTraceEvent = true
	}
}

type (
	// Option configures the logger with specific options.
	Option  func(*options)
	options struct {
		handler      slog.Handler
		sampler      func(context.Context) bool
		asTraceEvent bool
	}
)
