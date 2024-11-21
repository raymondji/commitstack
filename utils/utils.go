package utils

import (
	"errors"
	"sync"
)

func RunConcurrently[T any](functions []func() (T, error)) ([]T, error) {
	var wg sync.WaitGroup
	resultCh := make(chan T, len(functions))
	errCh := make(chan error, len(functions))

	// Run each function concurrently
	for _, fn := range functions {
		wg.Add(1)
		go func(fn func() (T, error)) {
			defer wg.Done()
			result, err := fn()
			if err != nil {
				errCh <- err
			} else {
				resultCh <- result
			}
		}(fn)
	}

	// Wait for all functions to complete
	wg.Wait()
	close(resultCh)
	close(errCh)

	// Collect results and errors
	var results []T
	var errs []error
	for result := range resultCh {
		results = append(results, result)
	}
	for err := range errCh {
		errs = append(errs, err)
	}

	return results, errors.Join(errs...)
}
