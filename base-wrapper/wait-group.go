package basewrapper

import (
	"iter"
	"sync"
)

func RunParallelRoutines[resChan chan resChanArg, resChanArg any](
	resultChan resChan,
	routineIter iter.Seq[func() resChanArg],
	collectorCb func(arg resChanArg) error,
) error {
	// Trigger sub routines
	waitGroup := sync.WaitGroup{}

	routineCount := 0
	for routine := range routineIter {
		routineCount++

		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()

			resultChan <- routine()
		}()
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
