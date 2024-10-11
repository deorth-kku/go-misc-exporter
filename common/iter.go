package common

type pair[K any, V any] struct {
	Key   K
	Value V
}

type PairSlice[K any, V any] []pair[K, V]

func (ps PairSlice[K, V]) Range(yield func(K, V) bool) {
	for _, pair := range ps {
		if !yield(pair.Key, pair.Value) {
			return
		}
	}
}
