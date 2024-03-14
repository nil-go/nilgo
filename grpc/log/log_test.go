// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log_test

import (
	"log/slog"
	"testing"

	"github.com/nil-go/sloth/rate"
	"github.com/nil-go/sloth/sampling"
	"github.com/stretchr/testify/assert"

	"github.com/nil-go/nilgo/grpc/log"
)

func TestServerOptions(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		handler     slog.Handler
	}{
		{
			description: "nil handler",
		},
		{
			description: "sampling handler",
			handler:     sampling.Handler{},
		},
		{
			description: "non sampling handler",
			handler:     rate.Handler{},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			options := log.ServerOptions(testcase.handler)
			assert.Len(t, options, 2)
		})
	}
}
