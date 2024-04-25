// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package gcp

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
// By default, it reads from environment variable "K_SERVICE" if it's running on GCP.
func WithService(service string) Option {
	return func(options *options) {
		options.service = service
	}
}

// WithVersion provides the GCP service version.
//
// By default, it reads from environment variable "K_REVISION" if it's running on GCP.
func WithVersion(version string) Option {
	return func(options *options) {
		options.version = version
	}
}

// WithLogOptions provides the gcp.Option(s) to configure the logger.
func WithLogOptions(opts gcp.Option) Option {
	return func(options *options) {
		options.logOpts = append(options.logOpts, opts)
	}
}

// WithOptions provides the function which returns multiple Option(s).
// It's useful while the Option needs to read configuration from config,
// since it defers the creation of Option(s) until the config is loaded.
func WithOptions(f func() []Option) Option {
	return func(options *options) {
		for _, opt := range f() {
			opt(options)
		}
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
		profilerOptions
	}
)
