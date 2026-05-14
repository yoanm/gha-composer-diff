package ddgha

import "iter"

func MapToCallbackIterator[K comparable, V any, R any](
	theMap map[K]V,
	callback func(key K, val V) R,
) iter.Seq[R] {
	return func(yield func(R) bool) {
		for key, value := range theMap {
			if !yield(callback(key, value)) {
				return
			}
		}
	}
}
