// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package client provides opinionated production-ready HTTP  client.
package client

import (
	"net"
	"net/http"
	"time"

	"github.com/nil-go/nilgo/http/internal"
)

// New creates a new http client with recommended production-ready settings.
func New(opts ...Option) *http.Client {
	option := &options{}
	for _, opt := range opts {
		opt(option)
	}
	if option.timeout == 0 {
		option.timeout = time.Second
	}
	if option.maxConnections == 0 {
		option.maxConnections = 100
	}

	transportTimeout := option.timeout / 5 //nolint:mnd
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
