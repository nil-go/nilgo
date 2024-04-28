// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal_test

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/nil-go/nilgo/http/internal"
	"github.com/nil-go/nilgo/http/internal/assert"
)

func TestRegisterUnixProtocol(t *testing.T) {
	t.Parallel()

	randBytes := make([]byte, 4) //nolint:makezero
	_, err := rand.Read(randBytes)
	assert.NoError(t, err)
	endpoint := "." + hex.EncodeToString(randBytes) + ".sock"
	defer func() {
		_ = os.Remove(endpoint)
	}()

	go func() {
		server := http.Server{
			Addr: endpoint,
			Handler: http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
				assert.Equal(t, "/", request.URL.Path)
			}),
			ReadTimeout: time.Second,
		}
		listener, serr := net.Listen("unix", endpoint)
		assert.NoError(t, serr)
		serr = server.Serve(listener)
		assert.NoError(t, serr)
	}()
	time.Sleep(time.Second) // Wait for server to start.

	internal.RegisterUnixProtocol(http.DefaultTransport.(*http.Transport))
	request, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "unix:"+endpoint+"/?query=val", nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(request)
	defer func() { _ = resp.Body.Close() }()
	assert.NoError(t, err)
}
