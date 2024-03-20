// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/nil-go/sloth/sampling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/nil-go/nilgo/grpc/log"
)

func TestBufferUnaryInterceptor(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		records     []slog.Record
		expected    string
	}{
		{
			description: "info only",
			records: []slog.Record{
				slog.NewRecord(time.Time{}, slog.LevelInfo, "info", 0),
			},
		},
		{
			description: "with error",
			records: []slog.Record{
				slog.NewRecord(time.Time{}, slog.LevelInfo, "info", 0),
				slog.NewRecord(time.Time{}, slog.LevelError, "error", 0),
			},
			expected: `level=INFO msg=info
level=ERROR msg=error
`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := sampling.New(
				slog.NewTextHandler(buf, &slog.HandlerOptions{}),
				func(context.Context) bool { return false },
			)

			resp, err := log.BufferUnaryInterceptor(context.Background(), nil, nil, func(ctx context.Context, _ any) (any, error) {
				for _, record := range testcase.records {
					_ = handler.Handle(ctx, record)
				}

				return "", nil
			})
			require.NoError(t, err)
			assert.Equal(t, "", resp)
			assert.Equal(t, testcase.expected, buf.String())
		})
	}
}

func TestBufferStreamInterceptor(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		records     []slog.Record
		expected    string
	}{
		{
			description: "info only",
			records: []slog.Record{
				slog.NewRecord(time.Time{}, slog.LevelInfo, "info", 0),
			},
		},
		{
			description: "with error",
			records: []slog.Record{
				slog.NewRecord(time.Time{}, slog.LevelInfo, "info", 0),
				slog.NewRecord(time.Time{}, slog.LevelError, "error", 0),
			},
			expected: `level=INFO msg=info
level=ERROR msg=error
`,
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := sampling.New(
				slog.NewTextHandler(buf, &slog.HandlerOptions{}),
				func(context.Context) bool { return false },
			)

			err := log.BufferStreamInterceptor(nil, serverStream{}, nil, func(_ any, stream grpc.ServerStream) error {
				for _, record := range testcase.records {
					_ = handler.Handle(stream.Context(), record)
				}

				return nil
			})
			require.NoError(t, err)
			assert.Equal(t, testcase.expected, buf.String())
		})
	}
}

type serverStream struct{ grpc.ServerStream }

func (s serverStream) Context() context.Context { return context.Background() }
