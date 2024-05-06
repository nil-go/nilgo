// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log

import (
	"context"
)

// WithSampler provides a sampler function which decides whether Info logs should write to output.
//
// By default, it disables simpling with nil sampler.
func WithSampler(sampler func(context.Context) bool) Option {
	return func(options *options) {
		options.sampler = sampler
	}
}

type (
	// Option configures the logger with specific options.
	Option  func(*options)
	options struct {
		sampler func(context.Context) bool
	}
)
