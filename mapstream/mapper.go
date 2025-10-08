package mapstream

import (
	"context"
)

func Mapper[I any, O any](
	ctx context.Context,
	inChan <-chan I,
	outChan chan<- O,
	f func(context.Context, I) (O, error),
) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case input, open := <-inChan:
			if !open {
				return nil
			}

			output, err := f(ctx, input)

			if err != nil {
				return err
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case outChan <- output:
			}
		}
	}

}
