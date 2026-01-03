package common

import (
	"context"
	"sync"
)

// ParallelResult holds the result of a parallel operation
type ParallelResult[R any] struct {
	Result R
	Err    error
}

// ParallelMap executes fn on each item in parallel with a worker pool
// Returns results in the same order as input items
func ParallelMap[T any, R any](ctx context.Context, items []T, fn func(context.Context, T) (R, error), workers int) ([]R, error) {
	if len(items) == 0 {
		return nil, nil
	}

	if workers <= 0 {
		workers = len(items)
	}
	if workers > len(items) {
		workers = len(items)
	}

	type indexedItem struct {
		index int
		item  T
	}

	type indexedResult struct {
		index  int
		result R
		err    error
	}

	jobs := make(chan indexedItem, len(items))
	results := make(chan indexedResult, len(items))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobs {
				select {
				case <-ctx.Done():
					results <- indexedResult{index: job.index, err: ctx.Err()}
				default:
					r, err := fn(ctx, job.item)
					results <- indexedResult{index: job.index, result: r, err: err}
				}
			}
		}()
	}

	// Send jobs
	for i, item := range items {
		jobs <- indexedItem{index: i, item: item}
	}
	close(jobs)

	// Wait for workers
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results in order
	ordered := make([]R, len(items))
	var firstErr error
	for r := range results {
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
		ordered[r.index] = r.result
	}

	return ordered, firstErr
}

// ParallelDo executes multiple functions in parallel and returns the first error
func ParallelDo(ctx context.Context, tasks ...func(context.Context) error) error {
	if len(tasks) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(tasks))

	for _, task := range tasks {
		wg.Add(1)
		go func(t func(context.Context) error) {
			defer wg.Done()
			if err := t(ctx); err != nil {
				errCh <- err
			}
		}(task)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	// Return first error if any
	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// FanOut executes a function concurrently for each item and collects results
// Simpler than ParallelMap, doesn't preserve order
func FanOut[T any, R any](ctx context.Context, items []T, fn func(context.Context, T) (R, error)) ([]R, error) {
	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		results  []R
		firstErr error
	)

	for _, item := range items {
		wg.Add(1)
		go func(i T) {
			defer wg.Done()
			r, err := fn(ctx, i)
			mu.Lock()
			defer mu.Unlock()
			if err != nil && firstErr == nil {
				firstErr = err
			}
			results = append(results, r)
		}(item)
	}

	wg.Wait()
	return results, firstErr
}
