// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log_test

import (
	"bytes"
	"context"
	"log/slog"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/nil-go/nilgo/grpc/log"
)

func TestRecoveryUnaryInterceptor(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		handler     grpc.UnaryHandler
		log         string
		err         string
	}{
		{
			description: "normal",
			handler: func(context.Context, any) (any, error) {
				return "", nil
			},
		},
		{
			description: "panic",
			handler: func(context.Context, any) (any, error) {
				panic("panic from handler")
			},
			log: `level=ERROR source=/recovery_test.go:39 msg="Panic Recovered" error="panic from handler"
`,
			err: "rpc error: code = Internal desc = ",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
				AddSource: true,
				ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
					if len(groups) == 0 && attr.Key == slog.TimeKey {
						return slog.Attr{}
					}

					return attr
				},
			})

			resp, err := log.RecoveryUnaryInterceptor(handler)(context.Background(), nil, nil, testcase.handler)
			if testcase.err == "" {
				require.NoError(t, err)
				assert.Equal(t, "", resp)
			} else {
				require.EqualError(t, err, testcase.err)
			}
			pwd, _ := os.Getwd()
			assert.Equal(t, testcase.log, strings.ReplaceAll(buf.String(), pwd, ""))
		})
	}
}

func TestRecoveryStreamInterceptor(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		description string
		handler     grpc.StreamHandler
		log         string
		err         string
	}{
		{
			description: "normal",
			handler: func(any, grpc.ServerStream) error {
				return nil
			},
		},
		{
			description: "panic",
			handler: func(any, grpc.ServerStream) error {
				panic("panic from handler")
			},
			log: `level=ERROR source=/recovery_test.go:96 msg="Panic Recovered" error="panic from handler"
`,
			err: "rpc error: code = Internal desc = ",
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
				AddSource: true,
				ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
					if len(groups) == 0 && attr.Key == slog.TimeKey {
						return slog.Attr{}
					}

					return attr
				},
			})

			err := log.RecoveryStreamInterceptor(handler)(nil, serverStream{}, nil, testcase.handler)
			if testcase.err == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testcase.err)
			}
			pwd, _ := os.Getwd()
			assert.Equal(t, testcase.log, strings.ReplaceAll(buf.String(), pwd, ""))
		})
	}
}
