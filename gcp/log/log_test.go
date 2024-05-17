// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nil-go/nilgo/gcp/log"
)

func TestLogHandler(t *testing.T) {
	t.Parallel()

	assert.NotNil(t, log.Logger(
		log.WithService("test"),
		log.WithVersion("dev"),
		log.WithProject("project"),
	))
}
