package ddgha

import (
	"iter"
	"sync"
)

func RunParallelRoutines[R any, C chan R](
	resultChan C,
	routineIter iter.Seq[func() R],
	collectorCb func(arg R) error,
) error {
	// Trigger sub routines
	waitGroup := sync.WaitGroup{}

	routineCount := 0
	for res := range routineIter {
		routineCount++

		waitGroup.Add(1)

		go func(res func() R) {
			defer waitGroup.Done()

			resultChan <- res()
		}(res)
	}

	// Handle collecting results from the channel
	defer close(resultChan)
	defer waitGroup.Wait()

	var firstError error

	resultCount := 0

	for result := range resultChan {
		if err := collectorCb(result); err != nil {
			// Manage only the first error encountered !
			// The next ones are likely to be a consequence of the first one anyway (e.g. context cancellation)
			firstError = err

			break
		}

		resultCount++
		if resultCount == routineCount {
			break
		}
	}

	return firstError
}
