package mapstream

import (
	"context"
)

func taskMapper[I any, O any](
	ctx context.Context,
	inChan <-chan Task[I],
	outChan chan<- Task[O],
	f func(context.Context, I) (O, error),
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case task, open := <-inChan:
			if !open {
				return nil
			}

			res, err := f(ctx, task.Value)

			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case outChan <- Task[O]{
				SequenceNumber: task.SequenceNumber,
				Value:          res,
			}:
			}
		}
	}

}
