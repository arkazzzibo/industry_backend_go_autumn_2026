package main

import (
	"context"
	"errors"
	"sync"
)

type Result[R any] struct {
	Value R
	Err   error
}

type Task[T any] struct {
	idx   int
	value T
}

var ErrInvalidWorkers = errors.New("workers must be positive")

func ParallelMap[T any, R any](ctx context.Context, workers int, in []T, fn func(context.Context, T) (R, error)) ([]Result[R], error) {
	if workers <= 0 {
		return nil, ErrInvalidWorkers
	}

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	if len(in) == 0 {
		return []Result[R]{}, nil
	}

	var wg sync.WaitGroup

	results := make([]Result[R], len(in))
	taskCh := make(chan Task[T])

	for i := 0; i < min(workers, len(in)); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				if ctx.Err() != nil {
					return
				}
				result, err := fn(ctx, task.value)
				if err != nil {
					results[task.idx] = Result[R]{
						Err: err, // убираю передачу значения тк нужно нулевое значение если ошибка
					}
					continue
				}
				results[task.idx] = Result[R]{
					Value: result,
				}
			}
		}()
	}

	go func() {
		defer close(taskCh)
		for idx, t := range in {
			select {
			case taskCh <- Task[T]{
				idx:   idx,
				value: t,
			}:
			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()

	if ctx.Err() != nil {
		return nil, ctx.Err()
	}

	return results, nil
}
