package mapstream

import (
	"context"

	"github.com/agnosticeng/concu/worker"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type MapStreamConfig struct {
	PoolSize          int
	SemaphoreWeight   int
	WorkerChanSize    int
	SequencerChanSize int
}

func MapStream[I any, O any](
	ctx context.Context,
	inChan chan I,
	outChan chan O,
	f func(context.Context, I) (O, error),
	conf MapStreamConfig,
) error {
	if conf.PoolSize <= 0 {
		conf.PoolSize = 1
	}

	if conf.SemaphoreWeight <= 0 {
		conf.SemaphoreWeight = 100
	}

	if conf.WorkerChanSize <= 0 {
		conf.WorkerChanSize = 10
	}

	if conf.SequencerChanSize <= 0 {
		conf.SequencerChanSize = 10
	}

	var (
		workerChan    = make(chan Task[I], conf.WorkerChanSize)
		sequencerChan = make(chan Task[O], conf.SequencerChanSize)
		sem           = semaphore.NewWeighted(int64(conf.SemaphoreWeight))
	)

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(
		func() error {
			return Stamper(groupCtx, inChan, workerChan, sem)
		},
	)

	group.Go(
		func() error {
			defer close(sequencerChan)

			return worker.RunN(
				groupCtx,
				conf.PoolSize,
				func(ctx context.Context, i int) func() error {
					return func() error {
						return Worker(ctx, workerChan, sequencerChan, f)
					}
				},
			)
		},
	)

	group.Go(
		func() error {
			return Sequencer(groupCtx, sequencerChan, outChan, sem)
		},
	)

	return group.Wait()
}
