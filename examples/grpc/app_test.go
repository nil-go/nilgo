// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

//go:build !race

package grpc_test

import (
	"context"
	"errors"
	"log"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection/grpc_reflection_v1"

	"github.com/nil-go/nilgo"
	ngrpc "github.com/nil-go/nilgo/grpc"
)

const endpoint = "unix:.grpc.sock"

func TestMain(m *testing.M) {
	server := grpc.NewServer()
	runner := nilgo.New(nilgo.WithPreRun(ngrpc.Run(server, ngrpc.WithAddress(endpoint))))

	if err := runner.Run(context.Background(), func(context.Context) error {
		if m.Run() != 0 {
			return errors.New("test failed")
		}

		return nil
	}); err != nil {
		log.Fatal(err)
	}

	if err := goleak.Find(); err != nil {
		log.Fatal(err)
	}
}

func TestHealthCheck(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	hcClient := grpc_health_v1.NewHealthClient(conn)
	hcResp, err := hcClient.Check(ctx, &grpc_health_v1.HealthCheckRequest{}, grpc.WaitForReady(true))
	require.NoError(t, err)
	require.Equal(t, grpc_health_v1.HealthCheckResponse_SERVING, hcResp.GetStatus())
}

func TestReflection(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.NewClient(
		endpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.NoError(t, err)
	defer func() {
		_ = conn.Close()
	}()

	refClient := grpc_reflection_v1.NewServerReflectionClient(conn)
	stream, err := refClient.ServerReflectionInfo(ctx, grpc.WaitForReady(true))
	require.NoError(t, err)
	require.NoError(t, stream.CloseSend())
}
