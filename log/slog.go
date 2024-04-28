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
	"context"
	"log/slog"

	"github.com/nil-go/sloth/otel"
	"github.com/nil-go/sloth/rate"
	"github.com/nil-go/sloth/sampling"
	"go.opentelemetry.io/otel/trace"
)

// New creates a new slog.Logger with the given Option(s).
func New(opts ...Option) *slog.Logger {
	option := options{}
	for _, opt := range opts {
		opt(&option)
	}
	if option.handler == nil {
		return slog.Default()
	}

	var handler slog.Handler = rate.New(option.handler)

	if option.asTraceEvent {
		// If the logger is configured to log as trace event, it disables sampling.
		// However, sampling handler still can buffer and logs if there is a error log,
		// or there is no valid trace context.
		handler = sampling.New(handler, func(ctx context.Context) bool {
			return !trace.SpanContextFromContext(ctx).IsValid()
		})
		handler = otel.New(handler, otel.WithRecordEvent(true))

		return slog.New(handler)
	}

	if option.sampler != nil {
		handler = sampling.New(handler, option.sampler)
	}

	return slog.New(handler)
}
