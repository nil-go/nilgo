// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package log provides a structured logging base on [slog].
//
// # Sampling
//
// It supports sampling Info logs to reduce the log volume using sampler provided by WithSampler.
// However, if there exists an Error log, it records buffered Info logs for same request event they are not sampled.
// It provides context for root cause analysis while Error happens.
//
// # Trace Integration
//
// It supports recording log records as trace span's events if it's enabled by WithLogAsTraceEvent.
// It could significantly reduce the log volume then cost as trace is priced by number of span.
//
// # Rate limiting
//
// Applications often experience runs of errors, either because of a bug or because of a misbehaving user.
// Logging errors is usually a good idea, but it can easily make this bad situation worse:
// not only is your application coping with a flood of errors, it's also spending extra CPU cycles and I/O
// logging those errors. Since writes are typically serialized, logging limits throughput when you need it most.
//
// Rate limiting fixes this problem by dropping repetitive log entries. Under normal conditions,
// your application writes out every entry. When similar entries are logged hundreds or thousands of times each second,
// though, it begins dropping duplicates to preserve throughput.
package log

import (
	"log/slog"

	"github.com/nil-go/sloth/otel"
	"github.com/nil-go/sloth/rate"
	"github.com/nil-go/sloth/sampling"
)

// New creates a new slog.Logger with the given handler and Option(s).
func New(handler slog.Handler, opts ...Option) *slog.Logger {
	if handler == nil {
		return slog.Default()
	}

	option := options{}
	for _, opt := range opts {
		opt(&option)
	}

	handler = rate.New(handler)
	if option.sampler != nil {
		handler = sampling.New(handler, option.sampler)
	}
	handler = otel.New(handler)

	return slog.New(handler)
}
