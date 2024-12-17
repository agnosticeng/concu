package mapstream

import (
	"context"
	"errors"
	"math"
	"time"

	"golang.org/x/sync/semaphore"
)

var ErrNoMoreSequenceNumber = errors.New("no more sequence number")

func Stamper[T any](
	ctx context.Context,
	inChan chan T,
	outChan chan Task[T],
	sem *semaphore.Weighted,
) error {
	defer close(outChan)

	var nextSequenceNumber uint64

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case item, open := <-inChan:
			if !open {
				return nil
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case outChan <- Task[T]{SequenceNumber: nextSequenceNumber, CreatedAt: time.Now(), Value: item}:
				nextSequenceNumber++

				if nextSequenceNumber == math.MaxUint64 {
					return ErrNoMoreSequenceNumber
				}
			}
		}
	}

}
