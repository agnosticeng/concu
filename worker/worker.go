package worker

import (
	"context"

	"golang.org/x/sync/errgroup"
)

type WorkerFactoryFunc func(ctx context.Context, i int) func() error

type OnExitFunc func()

func RunN(
	ctx context.Context,
	numWorkers int,
	f WorkerFactoryFunc,
) error {
	if numWorkers <= 0 {
		numWorkers = 1
	}

	var group, groupctx = errgroup.WithContext(ctx)

	for i := 0; i < numWorkers; i++ {
		group.Go(func() error {
			return f(groupctx, i)()
		})
	}

	return group.Wait()
}
