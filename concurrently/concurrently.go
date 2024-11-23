package concurrently

import (
	"errors"
	"sync"
)

func ForEach[T any, Result any](values []T, f func(v T) (Result, error)) ([]Result, error) {
	var wg sync.WaitGroup
	errs := make([]error, len(values))
	res := make([]Result, len(values))

	for i := range values {
		wg.Add(1)
		pin := i
		go func() {
			defer wg.Done()
			v := values[pin]
			res[pin], errs[pin] = f(v)
		}()
	}

	wg.Wait()

	var finalErrs []error
	var finalRes []Result
	for i := range errs {
		if errs[i] == nil {
			finalRes = append(finalRes, res[i])
		} else {
			finalErrs = append(finalErrs, errs[i])
		}
	}

	if len(finalErrs) > 0 {
		return nil, errors.Join(finalErrs...)
	}

	return finalRes, nil
}
