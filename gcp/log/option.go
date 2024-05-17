// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package log

import (
	"github.com/nil-go/sloth/gcp"
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

// WithOption provides the gcp.Option(s) to configure the logger.
func WithOption(opts ...gcp.Option) Option {
	return func(options *options) {
		options.logOpts = append(options.logOpts, opts...)
	}
}

type (
	// Option configures the GCP runtime with specific options.
	Option  func(*options)
	options struct {
		project string
		service string
		version string
		logOpts []gcp.Option
	}
)
