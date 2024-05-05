// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package client

import "time"

// WithTimeout provides the duration that timeout http request.
//
// By default, it has 1 second timeout.
func WithTimeout(timeout time.Duration) Option {
	return func(options *options) {
		options.timeout = timeout
	}
}

// WithMaxConnections provides the maximum number of idle (keep-alive) connections
// across all hosts.
//
// By default, it has 100 idle connections.
func WithMaxConnections(maxConnections uint) Option {
	return func(options *options) {
		options.maxConnections = maxConnections
	}
}

// WithUnixSocket enables the unix socket support for the http client.
func WithUnixSocket() Option {
	return func(options *options) {
		options.unixSocket = true
	}
}

type (
	// Option configures the http Client.
	Option  func(*options)
	options struct {
		timeout        time.Duration
		maxConnections uint
		unixSocket     bool
	}
)
