package hlpr

// GroupBy группирует значения массива ts по ключу, определяемый функцией key.
func GroupBy[K comparable, V any](ts []V, key func(v V) K) map[K][]V {
	res := make(map[K][]V)
	for _, t := range ts {
		if _, ok := res[key(t)]; !ok {
			res[key(t)] = make([]V, 0, 1)
		}
		res[key(t)] = append(res[key(t)], t)
	}
	return res
}
