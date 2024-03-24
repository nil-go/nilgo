// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

//go:build !race

package grpc_test

import (
	"context"
	"log/slog"
	"strings"
	"testing"

	"github.com/nil-go/konf"
	"github.com/nil-go/sloth/sampling"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/interop"
	"google.golang.org/grpc/interop/grpc_testing"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"

	ngrpc "github.com/nil-go/nilgo/grpc"
	pb "github.com/nil-go/nilgo/grpc/pb/nilgo/v1"
)

func TestRun(t *testing.T) {
	buf := new(strings.Builder)

	spanRecorder := tracetest.NewSpanRecorder()
	metricReader := metric.NewManualReader()
	testcases := []struct {
		description string
		server      func() *grpc.Server
		check       func(conn *grpc.ClientConn)
	}{
		{
			description: "nil server",
			server:      func() *grpc.Server { return nil },
		},
		{
			description: "organic server",
			server: func() *grpc.Server {
				server := grpc.NewServer()
				grpc_health_v1.RegisterHealthServer(server, health.NewServer())
				reflection.Register(server)

				return server
			},
		},
		{
			description: "default server",
			server: func() *grpc.Server {
				server := ngrpc.NewServer()
				grpc_testing.RegisterTestServiceServer(server, interop.NewTestServer())

				return server
			},
			check: func(conn *grpc.ClientConn) {
				ctx := context.Background()

				client := grpc_testing.NewTestServiceClient(conn)
				interop.DoEmptyUnaryCall(ctx, client)
				interop.DoLargeUnaryCall(ctx, client)
				interop.DoEmptyStream(ctx, client)
			},
		},
		{
			description: "default config service",
			server: func() *grpc.Server {
				return ngrpc.NewServer(ngrpc.WithConfigService())
			},
			check: func(conn *grpc.ClientConn) {
				client := pb.NewConfigServiceClient(conn)
				resp, err := client.Explain(context.Background(), &pb.ExplainRequest{Path: "user"})
				require.NoError(t, err)
				assert.NotEmpty(t, resp.GetExplanation())
			},
		},
		{
			description: "config service",
			server: func() *grpc.Server {
				return ngrpc.NewServer(ngrpc.WithConfigService(konf.New(), konf.New()))
			},
			check: func(conn *grpc.ClientConn) {
				client := pb.NewConfigServiceClient(conn)
				resp, err := client.Explain(context.Background(), &pb.ExplainRequest{Path: "user"})
				require.NoError(t, err)
				assert.Equal(t, "user has no configuration.\n\n\n-----\nuser has no configuration.\n\n", resp.GetExplanation())
			},
		},
		{
			description: "panic recovery",
			server: func() *grpc.Server {
				handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
					ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
						if len(groups) == 0 && attr.Key == slog.TimeKey {
							return slog.Attr{}
						}

						return attr
					},
				})
				server := ngrpc.NewServer(ngrpc.WithLogHandler(handler))
				grpc_testing.RegisterTestServiceServer(server, panicServer{interop.NewTestServer()})

				return server
			},
			check: func(conn *grpc.ClientConn) {
				client := grpc_testing.NewTestServiceClient(conn)
				resp, err := client.UnimplementedCall(context.Background(), &grpc_testing.Empty{})
				require.EqualError(t, err, "rpc error: code = Internal desc = ")
				assert.Nil(t, resp)
				assert.Equal(t, `level=ERROR msg="Panic Recovered" error="unimplemented panic"
`, buf.String())
				buf.Reset()
			},
		},
		{
			description: "with telemetry",
			server: func() *grpc.Server {
				tp := trace.NewTracerProvider(trace.WithSpanProcessor(spanRecorder))
				mp := metric.NewMeterProvider(metric.WithReader(metricReader))

				return ngrpc.NewServer(
					ngrpc.WithTelemetry(otelgrpc.WithTracerProvider(tp), otelgrpc.WithMeterProvider(mp)),
				)
			},
			check: func(*grpc.ClientConn) {
				spans := spanRecorder.Ended()
				assert.Len(t, spans, 1)
				assert.Equal(t, "grpc.health.v1.Health/Check", spans[0].Name())
				rm := metricdata.ResourceMetrics{}
				err := metricReader.Collect(context.Background(), &rm)
				require.NoError(t, err)
				assert.Len(t, rm.ScopeMetrics, 1)
				metrics := rm.ScopeMetrics[0].Metrics
				assert.Len(t, metrics, 5)
				for _, m := range metrics {
					assert.True(t, strings.HasPrefix(m.Name, "rpc.server."))
				}
			},
		},
		{
			description: "sampling handler",
			server: func() *grpc.Server {
				t.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
				var handler slog.Handler = slog.NewTextHandler(buf, &slog.HandlerOptions{
					ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
						if len(groups) == 0 && attr.Key == slog.TimeKey {
							return slog.Attr{}
						}

						return attr
					},
				})
				handler = sampling.New(handler, func(context.Context) bool { return false })

				return ngrpc.NewServer(ngrpc.WithLogHandler(handler))
			},
			check: func(*grpc.ClientConn) {
				assert.Empty(t, buf)
				buf.Reset()
			},
		},
		{
			description: "slog handler",
			server: func() *grpc.Server {
				t.Setenv("GRPC_GO_LOG_SEVERITY_LEVEL", "info")
				handler := slog.NewTextHandler(buf, &slog.HandlerOptions{
					ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
						if len(groups) == 0 && attr.Key == slog.TimeKey {
							return slog.Attr{}
						}

						return attr
					},
				})

				return ngrpc.NewServer(ngrpc.WithLogHandler(handler))
			},
			check: func(*grpc.ClientConn) {
				assert.NotEmpty(t, buf)
				buf.Reset()
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			endpoint := t.TempDir() + "/test.sock"
			go func() {
				err := ngrpc.Run(testcase.server(), "unix://"+endpoint, "")(ctx)
				assert.NoError(t, err)
			}()

			conn, err := grpc.DialContext(
				ctx, "unix://"+endpoint,
				grpc.WithBlock(),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			require.NoError(t, err)

			hcClient := grpc_health_v1.NewHealthClient(conn)
			hcResp, err := hcClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{})
			require.NoError(t, err)
			require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, hcResp.GetStatus())

			refClient := grpc_reflection_v1.NewServerReflectionClient(conn)
			stream, err := refClient.ServerReflectionInfo(ctx)
			require.NoError(t, err)
			require.NoError(t, stream.CloseSend())

			if testcase.check != nil {
				testcase.check(conn)
			}
		})
	}
}

type panicServer struct{ grpc_testing.TestServiceServer }

func (s panicServer) UnimplementedCall(context.Context, *grpc_testing.Empty) (*grpc_testing.Empty, error) {
	panic("unimplemented panic")
}
