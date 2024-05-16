// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package grpc_test

import (
	"context"
	"testing"

	"github.com/nil-go/konf"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	t.Parallel()

	testcases := []struct {
		description string
		server      func() *grpc.Server
		opts        []ngrpc.Option
		assertion   func(*grpc.ClientConn)
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
			assertion: func(conn *grpc.ClientConn) {
				ctx := context.Background()

				client := grpc_testing.NewTestServiceClient(conn)
				interop.DoEmptyUnaryCall(ctx, client)
				interop.DoLargeUnaryCall(ctx, client)
				interop.DoEmptyStream(ctx, client)
			},
		},
		{
			description: "default config service",
			server:      func() *grpc.Server { return ngrpc.NewServer() },
			opts:        []ngrpc.Option{ngrpc.WithConfigService()},
			assertion: func(conn *grpc.ClientConn) {
				client := pb.NewConfigServiceClient(conn)
				resp, err := client.Explain(context.Background(), &pb.ExplainRequest{Path: "user"})
				require.NoError(t, err)
				assert.NotEmpty(t, resp.GetExplanation())
			},
		},
		{
			description: "config service",
			server:      func() *grpc.Server { return ngrpc.NewServer() },
			opts:        []ngrpc.Option{ngrpc.WithConfigService(konf.New(), konf.New())},
			assertion: func(conn *grpc.ClientConn) {
				client := pb.NewConfigServiceClient(conn)
				resp, err := client.Explain(context.Background(), &pb.ExplainRequest{Path: "user"})
				require.NoError(t, err)
				assert.Equal(t, "user has no configuration.\n\n\n-----\nuser has no configuration.\n\n", resp.GetExplanation())
			},
		},
		{
			description: "panic recovery",
			server: func() *grpc.Server {
				server := ngrpc.NewServer()
				grpc_testing.RegisterTestServiceServer(server, panicServer{interop.NewTestServer()})

				return server
			},
			assertion: func(conn *grpc.ClientConn) {
				client := grpc_testing.NewTestServiceClient(conn)
				resp, err := client.UnimplementedCall(context.Background(), &grpc_testing.Empty{})
				require.EqualError(t, err, "rpc error: code = Internal desc = ")
				assert.Nil(t, resp)
			},
		},
	}

	for _, testcase := range testcases {
		testcase := testcase

		t.Run(testcase.description, func(t *testing.T) {
			t.Parallel()

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			endpoint := t.TempDir() + "/test.sock"
			go func() {
				err := ngrpc.Run(testcase.server(), append(testcase.opts, ngrpc.WithAddress("unix://"+endpoint))...)(ctx)
				assert.NoError(t, err)
			}()

			conn, err := grpc.NewClient(
				"unix://"+endpoint,
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

			if testcase.assertion != nil {
				testcase.assertion(conn)
			}
		})
	}
}

type panicServer struct{ grpc_testing.TestServiceServer }

func (s panicServer) UnimplementedCall(context.Context, *grpc_testing.Empty) (*grpc_testing.Empty, error) {
	panic("unimplemented panic")
}
