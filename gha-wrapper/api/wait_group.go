package api

import "sync"

func triggerAwaitedGoRoutines[resChan chan resChanArg, resChanArg any](
	group *sync.WaitGroup,
	resultChan resChan,
	routines func(yield func(func() resChanArg) bool),
) int {
	actualCount := 0
	for routine := range routines {
		actualCount++

		group.Add(1)

		go func() {
			defer group.Done()

			resultChan <- routine()
		}()
	}

	return actualCount
}

func collectAllAwaitedGoRoutines[resChan chan resChanArg, resChanArg any](
	group *sync.WaitGroup,
	maxIteration int,
	resultChan resChan,
	collector func(arg resChanArg) error,
) []error {
	defer close(resultChan)
	defer group.Wait()

	resultCount := 0

	errorList := []error{}

	for result := range resultChan {
		if err := collector(result); err != nil {
			errorList = append(errorList, err)
		}

		resultCount++
		if resultCount == maxIteration {
			break
		}
	}

	return errorList
}
