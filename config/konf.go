// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

// Package config provides an application configuration loader base on [konf].
//
// # Configuration Sources
//
// It loads configuration from the following sources,
// and each source takes precedence over the sources below it:
//
//   - config files specified by WithFS and WithFile.
//     WithFile also can be overridden by the environment variable `CONFIG_FILE`.
//     For example, if CONFIG_FILE = "f1, f2,f3", it will load f1, f2, and f3,
//     and each file takes precedence over the files before it.
//   - environment variables which matches the following pattern:
//     prefix + "_" + key, all in ALL CAPS.
//     For example, FOO_BAR is the name of environment variable for configuration `foo.bar`.
//
// [konf]: https://pkg.go.dev/github.com/nil-go/konf
package config

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/nil-go/konf"
	"github.com/nil-go/konf/provider/env"
	"github.com/nil-go/konf/provider/fs"
	"gopkg.in/yaml.v3"
)

// New creates a new konf.Config with the given Option(s).
func New(opts ...Option) (*konf.Config, error) {
	options := options{}
	for _, opt := range opts {
		opt(&options)
	}
	if files := konf.Get[[]string]("config.file"); len(files) > 0 {
		options.files = files
	}
	if len(options.files) == 0 {
		options.files = []string{"config/config.yaml"}
	}

	config := konf.New(options.opts...)
	for _, file := range options.files {
		if err := config.Load(fs.New(options.fs, file, fs.WithUnmarshal(yaml.Unmarshal))); err != nil {
			var e *os.PathError
			if !errors.As(err, &e) {
				return nil, fmt.Errorf("load config file %s: %w", file, err)
			}

			// Ignore not found error since config file is optional.
			slog.Warn("Config file not found.", "file", file)
		}
	}
	// Ignore error: env loader does not return error.
	_ = config.Load(env.New())
	if options.fn != nil {
		if err := options.fn(config); err != nil {
			return nil, err
		}
	}

	return config, nil
}
