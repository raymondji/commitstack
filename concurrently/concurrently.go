package concurrently

import (
	"errors"
	"sync"
)

func ForEach[T any, Result any](values []T, f func(v T) (Result, error)) ([]Result, error) {
	var wg sync.WaitGroup
	errCh := make(chan error, len(values))
	resCh := make(chan Result, len(values))

	for i := range values {
		wg.Add(1)
		pin := i
		go func() {
			defer wg.Done()
			v := values[pin]
			res, err := f(v)
			errCh <- err
			resCh <- res
		}()
	}

	wg.Wait()
	close(errCh)
	close(resCh)

	var errs []error
	var res []Result
	for i := 0; i < len(values); i++ {
		select {
		case r := <-resCh:
			res = append(res, r)
		case e := <-errCh:
			errs = append(errs, e)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return res, nil
}
