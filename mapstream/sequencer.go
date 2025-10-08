package mapstream

import (
	"context"
	"time"

	"golang.org/x/sync/semaphore"
)

func sequencer[T any](
	ctx context.Context,
	inChan <-chan Task[T],
	outChan chan<- T,
	sem *semaphore.Weighted,
) error {
	var (
		buf                Tasks[T]
		nextSequenceNumber uint64
	)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(100 * time.Millisecond):

		case job, open := <-inChan:
			if !open {
				return nil
			}

			buf.Insert(job)

			if job.SequenceNumber != nextSequenceNumber {
				break
			}

			for {
				if len(buf) == 0 {
					break
				}

				job = buf[0]

				if job.SequenceNumber != nextSequenceNumber {
					break
				}

				select {
				case <-ctx.Done():
					return ctx.Err()
				case outChan <- job.Value:
					sem.Release(1)
				}

				nextSequenceNumber++
				buf = buf[1:]
			}
		}
	}
}
