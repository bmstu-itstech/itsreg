package hlpr

// GroupBy группирует значения массива ts по ключу, определяемый функцией key.
func GroupBy[K comparable, V any](ts []V, key func(v V) K) map[K][]V {
	res := make(map[K][]V)
	for _, t := range ts {
		k := key(t)
		if _, ok := res[k]; !ok {
			res[k] = make([]V, 0, 1)
		}
		res[k] = append(res[k], t)
	}
	return res
}
