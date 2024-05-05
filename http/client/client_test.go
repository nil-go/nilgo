// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package client_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	nhttp "github.com/nil-go/nilgo/http/client"
	"github.com/nil-go/nilgo/http/internal/assert"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		opts        []nhttp.Option
		assertion   func(*http.Client)
	}{
		{
			description: "default",
			assertion: func(client *http.Client) {
				client.Timeout = time.Second
				transport, ok := client.Transport.(*http.Transport)
				assert.Equal(t, true, ok)
				assert.Equal(t, 200*time.Millisecond, transport.TLSHandshakeTimeout)
				assert.Equal(t, true, transport.ForceAttemptHTTP2)
				assert.Equal(t, 100, transport.MaxIdleConns)
				assert.Equal(t, 100, transport.MaxIdleConnsPerHost)
				assert.Equal(t, 200*time.Millisecond, transport.TLSHandshakeTimeout)

				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "unix://.test.sock", nil)
				assert.NoError(t, err)
				_, err = client.Do(req) //nolint:bodyclose
				assert.EqualError(t, err, `Get "unix://.test.sock": unsupported protocol scheme "unix"`)
			},
		},
		{
			description: "with timeout",
			opts: []nhttp.Option{
				nhttp.WithTimeout(10 * time.Second),
			},
			assertion: func(client *http.Client) {
				client.Timeout = time.Second
				transport, ok := client.Transport.(*http.Transport)
				assert.Equal(t, true, ok)
				assert.Equal(t, 2*time.Second, transport.TLSHandshakeTimeout)
				assert.Equal(t, 2*time.Second, transport.TLSHandshakeTimeout)
			},
		},
		{
			description: "with max connections",
			opts: []nhttp.Option{
				nhttp.WithMaxConnections(10),
			},
			assertion: func(client *http.Client) {
				client.Timeout = time.Second
				transport, ok := client.Transport.(*http.Transport)
				assert.Equal(t, true, ok)
				assert.Equal(t, 10, transport.MaxIdleConns)
				assert.Equal(t, 10, transport.MaxIdleConnsPerHost)
			},
		},
		{
			description: "with unix socket",
			opts: []nhttp.Option{
				nhttp.WithUnixSocket(),
			},
			assertion: func(client *http.Client) {
				req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "unix://.test.sock", nil)
				assert.NoError(t, err)
				_, err = client.Do(req) //nolint:bodyclose
				assert.EqualError(t, err, `Get "http://.test.sock": dial unix .test.sock: connect: no such file or directory`)
			},
		},
	}

	for _, testcase := range testcases {
		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			client := nhttp.New(testcase.opts...)
			testcase.assertion(client)
		})
	}
}
