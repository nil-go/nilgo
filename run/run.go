// Copyright (c) 2024 The nilgo authors
// Use of this source code is governed by a MIT license found in the LICENSE file.

package run

import (
	"context"
	"errors"
	"sync"
)

// parallel executes the given runs in parallel.
//
// It blocks until all runs complete or ctx is done, then
// returns the first non-nil error if received from any run.
func parallel(ctx context.Context, runs ...func(context.Context) error) error {
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
