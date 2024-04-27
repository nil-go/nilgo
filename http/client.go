// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package http

import (
	"net"
	"net/http"
	"time"

	"github.com/nil-go/nilgo/http/internal"
)

// NewClient creates a new http client with recommended production-ready settings.
func NewClient(opts ...ClientOption) *http.Client {
	option := &clientOptions{}
	for _, opt := range opts {
		opt(option)
	}
	if option.timeout == 0 {
		option.timeout = time.Second
	}
	if option.maxConnections == 0 {
		option.maxConnections = 100
	}

	transportTimeout := option.timeout / 5 //nolint:gomnd
	transport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout: transportTimeout,
		}).DialContext,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        int(option.maxConnections),
		MaxIdleConnsPerHost: int(option.maxConnections),
		TLSHandshakeTimeout: transportTimeout,
	}
	if option.unixSocket {
		internal.RegisterUnixProtocol(transport)
	}

	return &http.Client{
		Transport: transport,
		Timeout:   option.timeout,
	}
}

// WithClientTimeout provides the duration that timeout http request.
//
// By default, it has 1 second timeout.
func WithClientTimeout(timeout time.Duration) ClientOption {
	return func(options *clientOptions) {
		options.timeout = timeout
	}
}

// WithClientMaxConnections provides the maximum number of idle (keep-alive) connections
// across all hosts.
//
// By default, it has 100 idle connections.
func WithClientMaxConnections(maxConnections uint) ClientOption {
	return func(options *clientOptions) {
		options.maxConnections = maxConnections
	}
}

// WithClientUnixSocket enables the unix socket support for the http client.
func WithClientUnixSocket() ClientOption {
	return func(options *clientOptions) {
		options.unixSocket = true
	}
}

type (
	// ClientOption configures the http Client.
	ClientOption  func(*clientOptions)
	clientOptions struct {
		timeout        time.Duration
		maxConnections uint
		unixSocket     bool
	}
)
