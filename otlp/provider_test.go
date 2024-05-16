// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package otlp_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/nil-go/nilgo/otlp"
)

func TestTraceProvider(t *testing.T) {
	t.Parallel()

	provider, err := otlp.TraceProvider()
	require.NoError(t, err)
	assert.NotNil(t, provider)
}

func TestMeterProvider(t *testing.T) {
	t.Parallel()

	provider, err := otlp.MeterProvider()
	require.NoError(t, err)
	assert.NotNil(t, provider)
}
