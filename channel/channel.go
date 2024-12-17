package channel

import (
	"context"
	"errors"
	"reflect"
	"time"
)

func ReadBatch[T any](
	ctx context.Context,
	c chan T,
	maxSize int,
	maxWait time.Duration,
) ([]T, bool, time.Duration, error) {
	var (
		batch     = make([]T, 0, maxSize)
		waitChan  <-chan time.Time
		waitBegin = time.Now()
	)

	for {
		select {
		case <-ctx.Done():
			return batch, false, time.Since(waitBegin), ctx.Err()

		case <-waitChan:
			return batch, false, time.Since(waitBegin), nil

		case item, open := <-c:
			if !open {
				return batch, true, time.Since(waitBegin), nil
			}

			batch = append(batch, item)

			if len(batch) == maxSize {
				return batch, false, time.Since(waitBegin), nil
			}

			if waitChan == nil {
				waitBegin = time.Now()
				waitChan = time.After(maxWait)
			}
		}
	}
}

func Send[T any](ctx context.Context, c chan T, v T) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case c <- v:
		return nil
	}
}

func SendWithTimeout[T any](ctx context.Context, c chan T, v T, d time.Duration) error {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return Send(ctx, c, v)
}

func Receive[T any](ctx context.Context, c chan T) (T, bool, error) {
	var zero T

	select {
	case <-ctx.Done():
		return zero, false, ctx.Err()
	case v, open := <-c:
		return v, open, nil
	}
}

func ReceiveWithTimeout[T any](ctx context.Context, c chan T, d time.Duration) (T, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, d)
	defer cancel()
	return Receive(ctx, c)
}

func MakeSice[T any](length int, chanSize int) []chan T {
	var s = make([]chan T, length)

	for i := 0; i < length; i++ {
		s[i] = make(chan T, chanSize)
	}

	return s
}

func Select[T any](chans []<-chan T) (int, T, error) {
	var (
		zero  T
		cases = make([]reflect.SelectCase, len(chans))
	)

	for i, ch := range chans {
		cases[i] = reflect.SelectCase{Dir: reflect.SelectRecv, Chan: reflect.ValueOf(ch)}
	}

	chosen, value, ok := reflect.Select(cases)

	if !ok {
		return chosen, zero, errors.New("channel closed")
	}

	return chosen, value.Interface().(T), nil
}
