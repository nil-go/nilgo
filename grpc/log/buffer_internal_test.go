// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log

import (
	"context"
	"log/slog"
	"testing"

	"github.com/nil-go/sloth/sampling"
	"github.com/stretchr/testify/assert"
)

func TestIsSamplingHandler(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		handler     slog.Handler
		expected    bool
	}{
		{
			description: "nil handler",
		},
		{
			description: "sampling handler",
			handler:     sampling.Handler{},
			expected:    true,
		},
		{
			description: "sampling handler (pointer)",
			handler:     &sampling.Handler{},
			expected:    true,
		},
		{
			description: "embed sampling handler",
			handler:     handlerEmbed{sampling.Handler{}},
			expected:    true,
		},
		{
			description: "embed sampling handler (pointer)",
			handler:     &handlerEmbed{sampling.Handler{}},
			expected:    true,
		},
		{
			description: "nested sampling handler",
			handler:     handlerWrapper{handler: sampling.Handler{}},
			expected:    true,
		},
		{
			description: "nested sampling handler (pointer)",
			handler:     &handlerWrapper{handler: sampling.Handler{}},
			expected:    true,
		},
		{
			description: "deep nested sampling handler",
			handler:     handlerWrapper{handler: handlerWrapper{handler: sampling.Handler{}}},
			expected:    true,
		},
		{
			description: "non sampling handler",
			handler:     slogDiscard{},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase
		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, testcase.expected, isSamplingHandler(testcase.handler))
		})
	}
}

type handlerWrapper struct {
	slogDiscard
	handler slog.Handler
}

type slogDiscard struct{}

func (slogDiscard) Enabled(context.Context, slog.Level) bool  { return false }
func (slogDiscard) Handle(context.Context, slog.Record) error { return nil }
func (slogDiscard) WithAttrs([]slog.Attr) slog.Handler        { return slogDiscard{} }
func (slogDiscard) WithGroup(string) slog.Handler             { return slogDiscard{} }

type handlerEmbed struct {
	slog.Handler
}
