// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/nil-go/nilgo/internal/assert"
	"github.com/nil-go/nilgo/run"
)

func TestWithCloser(t *testing.T) {
	testcases := []struct {
		description string
		run         func(ctx context.Context) error
		closer      func() error
		err         string
	}{
		{
			description: "no error",
			run: func(context.Context) error {
				return nil
			},
			closer: func() error {
				return nil
			},
		},
		{
			description: "run error",
			run: func(context.Context) error {
				return errors.New("run error")
			},
			err: "run error",
		},
		{
			description: "closer error",
			closer: func() error {
				return errors.New("closer error")
			},
			err: "closer error",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			err := run.WithCloser(testcase.run, testcase.closer)(ctx)
			if testcase.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, testcase.err)
			}
		})
	}
}
