// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run

import (
	"context"
	"errors"
	"sync"
)

// Parallel executes the given runs in parallel.
//
// It blocks until all runs complete or ctx is done, then
// returns the first non-nil error if received from any run.
func Parallel(ctx context.Context, runs ...func(context.Context) error) error {
	ctx, cancel := context.WithCancelCause(ctx)
	defer cancel(nil)

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(runs))
	for _, run := range runs {
		run := run
		go func() {
			defer waitGroup.Done()

			if err := run(ctx); err != nil {
				cancel(err)
			}
		}()
	}
	waitGroup.Wait()

	if err := context.Cause(ctx); err != nil && !errors.Is(err, ctx.Err()) {
		return err //nolint:wrapcheck
	}

	return nil
}

// WithCloser wraps the given run with a dedicated closer function,
// which should cause the run returns when closer function executes.
//
// closer function executes even if run function returns non-nil error.
// It guarantees both run and closer functions complete
// and return the first non-nil error if any.
func WithCloser(run func(context.Context) error, closer func() error) func(context.Context) error {
	if run == nil {
		run = func(context.Context) error { return nil }
	}
	if closer == nil {
		closer = func() error { return nil }
	}

	return func(ctx context.Context) error {
		ctx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)

		var waitGroup sync.WaitGroup
		waitGroup.Add(1)
		go func() {
			defer waitGroup.Done()

			if err := run(ctx); err != nil {
				cancel(err)
			}
		}()
		<-ctx.Done()

		err := closer()
		waitGroup.Wait()

		if e := context.Cause(ctx); e != nil && !errors.Is(e, ctx.Err()) {
			return e //nolint:wrapcheck
		}

		return err
	}
}
