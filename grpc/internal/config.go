// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package internal

import (
	"context"
	"strings"

	"github.com/nil-go/konf"

	pb "github.com/nil-go/nilgo/grpc/pb/nilgo/v1"
)

// ConfigServiceServer is an implementation of [pb.ConfigServiceServer].
type ConfigServiceServer struct {
	pb.UnimplementedConfigServiceServer

	configs []*konf.Config
}

// NewConfigServiceServer creates a new ConfigServiceServer with the provided configs.
func NewConfigServiceServer(configs []*konf.Config) *ConfigServiceServer {
	return &ConfigServiceServer{configs: configs}
}

func (c ConfigServiceServer) Explain(_ context.Context, request *pb.ExplainRequest) (*pb.ExplainResponse, error) {
	if len(c.configs) == 0 {
		return &pb.ExplainResponse{Explanation: konf.Explain(request.GetPath())}, nil
	}

	path := request.GetPath()
	var explanation strings.Builder
	for _, config := range c.configs {
		if explanation.Len() > 0 {
			explanation.WriteString("\n-----\n")
		}
		explanation.WriteString(config.Explain(path))
	}

	return &pb.ExplainResponse{Explanation: explanation.String()}, nil
}
