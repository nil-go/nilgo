// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package http_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	nhttp "github.com/nil-go/nilgo/http"
	"github.com/nil-go/nilgo/http/internal/assert"
)

//nolint:gosec
func TestRun(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		server      func() *http.Server
		opts        []nhttp.Option
		assertion   func(string)
	}{
		{
			description: "nil server",
			server:      func() *http.Server { return nil },
			assertion: func(endpoint string) {
				request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
				assert.NoError(t, err)
				resp, err := http.DefaultClient.Do(request)
				assert.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				assert.Equal(t, http.StatusNotFound, resp.StatusCode)
			},
		},
		{
			description: "organic server",
			server: func() *http.Server {
				return &http.Server{
					Handler: http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
						panic("test")
					}),
				}
			},
			assertion: func(endpoint string) {
				request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
				assert.NoError(t, err)
				resp, err := http.DefaultClient.Do(request)
				assert.NoError(t, err)
				assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
				bytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				assert.Equal(t, "Internal Server Error\n", string(bytes))
			},
		},
		{
			description: "with timeout",
			server: func() *http.Server {
				return &http.Server{
					Handler: http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
						timer := time.NewTimer(2 * time.Second)
						defer timer.Stop()

						select {
						case <-r.Context().Done():
						case <-timer.C:
						}
					}),
				}
			},
			opts: []nhttp.Option{
				nhttp.WithTimeout(time.Second),
			},
			assertion: func(endpoint string) {
				request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint, nil)
				assert.NoError(t, err)
				resp, err := http.DefaultClient.Do(request)
				assert.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
			},
		},
		{
			description: "with config service",
			server: func() *http.Server {
				return &http.Server{}
			},
			opts: []nhttp.Option{
				nhttp.WithConfigService(),
			},
			assertion: func(endpoint string) {
				request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, endpoint+"/_config/_not_found", nil)
				assert.NoError(t, err)
				resp, err := http.DefaultClient.Do(request)
				assert.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				assert.Equal(t, http.StatusOK, resp.StatusCode)
				bytes, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)
				assert.Equal(t, "_not_found has no configuration.\n\n", string(bytes))
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			randBytes := make([]byte, 4) //nolint:makezero
			_, err := rand.Read(randBytes)
			assert.NoError(t, err)
			endpoint := "." + hex.EncodeToString(randBytes) + ".sock"
			defer func() {
				_ = os.Remove(endpoint)
			}()
			go func() {
				err := nhttp.Run(testcase.server(), append(testcase.opts, nhttp.WithAddress("unix:"+endpoint))...)(ctx)
				assert.NoError(t, err)
			}()
			time.Sleep(100 * time.Millisecond) // Wait for server to start.

			if testcase.assertion != nil {
				testcase.assertion("unix:" + endpoint)
			}
		})
	}
}
