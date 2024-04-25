// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package gcp provides customization for application runs on GCP.
//
// It integrates with following GCP services:
//   - [Cloud Logging]
//   - [Cloud Error Reporting]
//   - [Cloud Profiler]
//
// [Cloud Logging]: https://cloud.google.com/logging
// [Cloud Error Reporting]: https://cloud.google.com/error-reporting
// [Cloud Profiler]: https://cloud.google.com/profiler
package gcp

import (
	"os"

	"cloud.google.com/go/compute/metadata"
	"github.com/nil-go/sloth/gcp"
)

// Options provides the [nilgo.Run] options for application runs on GCP.
//
// By default, only logging and error reporting are configured.
// Profiler need to enable explicitly
// using corresponding Option(s).
func Options(opts ...Option) []any {
	option := options{}
	for _, opt := range opts {
		opt(&option)
	}
	if option.project == "" {
		option.project, _ = metadata.ProjectID()
	}
	if option.service == "" {
		option.service = os.Getenv("K_SERVICE")
	}
	if option.version == "" {
		option.version = os.Getenv("K_REVISION")
	}

	appOpts := []any{
		gcp.New(append(option.logOpts, gcp.WithErrorReporting(option.service, option.version))...),
	}
	if option.project == "" {
		return appOpts
	}

	if runner := profile(&option); runner != nil {
		appOpts = append(appOpts, runner)
	}

	return appOpts
}
