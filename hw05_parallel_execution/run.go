package hw05parallelexecution

import (
	"errors"
	"sync"
	"sync/atomic"
)

var ErrErrorsLimitExceeded = errors.New("errors limit exceeded")

type Task func() error

// Run starts tasks in n goroutines and stops its work when receiving m errors from tasks.
func Run(tasks []Task, n, m int) error {
	if n <= 0 || m <= 0 {
		return ErrErrorsLimitExceeded
	}

	taskCh := make(chan Task, len(tasks))
	for _, task := range tasks {
		taskCh <- task
	}
	close(taskCh)

	var wg sync.WaitGroup
	var errorCount int64 // Счетчик ошибок

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for task := range taskCh {
				if atomic.LoadInt64(&errorCount) >= int64(m) {
					return
				}

				if err := task(); err != nil {
					atomic.AddInt64(&errorCount, 1)
				}
			}
		}()
	}

	wg.Wait()

	if atomic.LoadInt64(&errorCount) >= int64(m) {
		return ErrErrorsLimitExceeded
	}

	return nil
}
