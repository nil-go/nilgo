// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package gcp

import (
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/nil-go/sloth/gcp"
	"google.golang.org/api/option"
)

// WithProject provides the GCP project ID.
//
// By default, it uses the project ID from the metadata server if it's running on GCP.
func WithProject(project string) Option {
	return func(options *options) {
		options.project = project
	}
}

// WithService provides the GCP service name.
//
// By default, it reads from environment variable "K_SERVICE" if it's running on Google Cloud Run.
func WithService(service string) Option {
	return func(options *options) {
		options.service = service
	}
}

// WithVersion provides the GCP service version.
//
// By default, it reads from environment variable "K_REVISION" if it's running on Google Cloud Run.
func WithVersion(version string) Option {
	return func(options *options) {
		options.version = version
	}
}

// WithLog provides the gcp.Option(s) to configure the logger.
func WithLog(opts ...gcp.Option) Option {
	return func(options *options) {
		if options.logOpts == nil {
			options.logOpts = []gcp.Option{}
		}
		options.logOpts = append(options.logOpts, opts...)
	}
}

// WithProfiler provides the option.ClientOption(s) to configure the profiler.
func WithProfiler(opts ...option.ClientOption) Option {
	return func(options *options) {
		if options.profilerOpts == nil {
			options.profilerOpts = []option.ClientOption{}
		}
		options.profilerOpts = append(options.profilerOpts, opts...)
	}
}

type (
	// Option configures the GCP runtime with specific options.
	Option  func(*options)
	options struct {
		project string
		service string
		version string

		logOpts      []gcp.Option
		profilerOpts []option.ClientOption
	}
)

func (o *options) apply(opts []Option) {
	for _, opt := range opts {
		opt(o)
	}

	// Get service and version from Google Cloud Run environment variables.
	if o.service == "" {
		o.service = os.Getenv("K_SERVICE")
	}
	if o.version == "" {
		o.version = os.Getenv("K_REVISION")
	}
	if o.project == "" {
		o.project, _ = metadata.ProjectID()
	}
}
