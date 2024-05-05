// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package config

import (
	"io/fs"

	"github.com/nil-go/konf"
)

// WithFile explicitly provides the config file paths,
// and each file takes precedence over the files before it.
//
// By default, it uses "config/config.yaml".
// It can be overridden by the environment variable "CONFIG_FILE".
func WithFile(files ...string) Option {
	return func(options *options) {
		options.files = append(options.files, files...)
	}
}

// WithFS provides the fs.FS to load the config files from.
//
// By default, it uses OS file system under the current directory.
func WithFS(fs fs.FS) Option {
	return func(options *options) {
		options.fs = fs
	}
}

// WithOption provides the konf.Option to customize the config.
func WithOption(opts ...konf.Option) Option {
	return func(options *options) {
		options.opts = append(options.opts, opts...)
	}
}

type (
	// Option configures the config with specific options.
	Option  func(*options)
	options struct {
		opts  []konf.Option
		files []string
		fs    fs.FS
	}
)
