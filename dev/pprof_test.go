// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

//go:build !race

package dev_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/nil-go/nilgo/dev"
	"github.com/nil-go/nilgo/internal/assert"
)

func TestPProf(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		assert.NoError(t, dev.Pprof(ctx))
	}()
	time.Sleep(100 * time.Millisecond) // wait for pprof server to start.

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:6060/debug/pprof/", nil)
	assert.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	assert.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
