package salad

func SearchBinary[T any](arr []T, value T, less func(x, y T) bool) int {
	low, high := 0, len(arr)-1
	for low <= high {
		mid := low + (high-low)/2
		if less(arr[mid], value) {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}
	return low
}

func InsertSorted[T any](arr []T, value T, less func(x, y T) bool) []T {
	idx := SearchBinary(arr, value, less)
	arr = append(arr, value) // любое значение типа T
	copy(arr[idx+1:], arr[idx:])
	arr[idx] = value
	return arr
}
