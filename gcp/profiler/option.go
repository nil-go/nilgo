// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

//nolint:ireturn
package profiler

import (
	"cloud.google.com/go/profiler"
	"google.golang.org/api/option"
	"google.golang.org/api/option/internaloption"
)

// WithProject provides the GCP project ID.
func WithProject(project string) option.ClientOption {
	return &optionFunc{
		fn: func(config *profiler.Config) {
			config.ProjectID = project
		},
	}
}

// WithService provides the GCP service name.
func WithService(service string) option.ClientOption {
	return &optionFunc{
		fn: func(config *profiler.Config) {
			config.Service = service
		},
	}
}

// WithVersion provides the GCP service version.
func WithVersion(version string) option.ClientOption {
	return &optionFunc{
		fn: func(config *profiler.Config) {
			config.ServiceVersion = version
		},
	}
}

// WithMutexProfiling enables mutex profiling.
func WithMutexProfiling() option.ClientOption {
	return &optionFunc{
		fn: func(config *profiler.Config) {
			config.MutexProfiling = true
		},
	}
}

// Option configures the profiler client.
type Option = option.ClientOption

type optionFunc struct {
	internaloption.EmbeddableAdapter
	fn func(*profiler.Config)
}
