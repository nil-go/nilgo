// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package profiler_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nil-go/nilgo/gcp/profiler"
)

func TestProfiler(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, profiler.Run(
		profiler.WithService("test"),
		profiler.WithVersion("dev"),
		profiler.WithProject("project"),
		profiler.WithMutexProfiling(),
	))
}
