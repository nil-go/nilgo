// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package gcp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nil-go/nilgo/gcp"
	"github.com/nil-go/nilgo/gcp/profiler"
)

func TestLogHandler(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, gcp.Logger(
		gcp.WithService("test"),
		gcp.WithVersion("dev"),
		gcp.WithProject("project"),
	))
}

func TestProfiler(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, gcp.Profiler(
		gcp.WithService("test"),
		gcp.WithVersion("dev"),
		gcp.WithProject("project"),
		gcp.WithProfiler(profiler.WithMutexProfiling()),
	))
}
