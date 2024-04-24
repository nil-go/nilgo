// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package nilgo

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nil-go/konf"

	"github.com/nil-go/nilgo/config"
	"github.com/nil-go/nilgo/log"
	"github.com/nil-go/nilgo/run"
)

func Run(args ...any) {
	if err := runInternal(args); err != nil {
		slog.Error("Panic due to unrecoverable error.", "error", err)
		panic(err.Error()) // It's unrecoverable.
	}
}

func runInternal(args []any) error {
	var (
		configOpts []config.Option
		logOpts    []log.Option
		runners    []func(context.Context) error
	)

	for _, arg := range args {
		switch opt := arg.(type) {
		case config.Option:
			configOpts = append(configOpts, opt)
		case log.Option:
			logOpts = append(logOpts, opt)
		case func(context.Context) error:
			runners = append(runners, opt)
		default:
			return fmt.Errorf("unknown argument type: %T", opt) //nolint:goerr113
		}
	}

	cfg, err := config.Init(configOpts...)
	if err != nil {
		return fmt.Errorf("init config: %w", err)
	}
	konf.SetDefault(cfg)

	logHandler := log.Init(logOpts...)
	if logHandler != nil {
		slog.SetDefault(slog.New(logHandler))
	}

	runner := run.New(run.WithPreRun(cfg.Watch))
	if err := runner.Run(context.Background(), runners...); err != nil {
		return fmt.Errorf("run: %w", err)
	}

	return nil
}
