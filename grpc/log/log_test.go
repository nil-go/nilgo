// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

//go:build !race

package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/nil-go/nilgo/grpc/log"
)

func TestServerOptions(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
	}{
		{
			description: "nil handler",
		},
		{
			description: "sampling handler",
		},
		{
			description: "non sampling handler",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			options := log.ServerOptions()
			assert.Len(t, options, 2)
		})
	}
}
