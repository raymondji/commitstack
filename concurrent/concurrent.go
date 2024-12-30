package concurrent

import (
	"context"

	"golang.org/x/sync/errgroup"
)

func ForEach[T any](
	ctx context.Context,
	values []T,
	f func(ctx context.Context, v T) error,
) error {
	_, err := Map(ctx, values, func(ctx context.Context, v T) (struct{}, error) {
		return struct{}{}, f(ctx, v)
	})
	return err
}

func Map[T any, Result any](
	ctx context.Context,
	values []T,
	f func(ctx context.Context, v T) (Result, error),
) ([]Result, error) {
	g, ctx := errgroup.WithContext(ctx)
	res := make([]Result, len(values))

	for i := range values {
		pin := i
		g.Go(func() error {
			var err error
			res[pin], err = f(ctx, values[pin])
			return err
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return res, nil
}
