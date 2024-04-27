// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal

import (
	"context"
	"net"
	"net/http"
	"strings"
)

const unix = "unix"

// RegisterUnixProtocol registers a protocol handler to the provided http.Transport
// that can server requests to Unix domain sockets via the "unix" schemes.
//
// Request URLs should have the following form:
//
//	unix:socket_file/request/path?query=val&...
//
// The registered transport is based on a clone of the provided transport,
// and uses the same configuration: timeouts, TLS settings, and so on.
// Connection pooling should also work as expected.
func RegisterUnixProtocol(transport *http.Transport) {
	defer func() {
		// Ignore panic from RegisterProtocol while protocol is already registered.
		_ = recover()
	}()

	uTransport := newUnixTransport(transport.Clone())
	transport.RegisterProtocol(unix, uTransport)
	transport.RegisterProtocol(unix+"s", uTransport)
}

type unixTransport struct {
	*http.Transport
}

func newUnixTransport(transport *http.Transport) unixTransport {
	dialContext := transport.DialContext
	if dialContext == nil {
		if t, ok := http.DefaultTransport.(*http.Transport); ok {
			dialContext = t.DialContext
		}
	}

	transport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
		host, _, err := net.SplitHostPort(addr)
		if err != nil {
			host = addr
		} else {
			network = unix
		}

		return dialContext(ctx, network, host)
	}

	return unixTransport{transport}
}

func (u unixTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.HasPrefix(req.URL.Scheme, unix) {
		req.URL.Scheme = strings.Replace(req.URL.Scheme, "unix", "http", 1)
		if req.URL.Host == "" && req.URL.Opaque != "" {
			req.URL.Host, req.URL.Path, _ = strings.Cut(req.URL.Opaque, "/")
			req.URL.Opaque = ""
		}
	}

	return u.Transport.RoundTrip(req) //nolint:wrapcheck
}
