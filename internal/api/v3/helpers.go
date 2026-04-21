package apiv3

func derefOrNilSlice[T any](ptr *[]T) []T {
	if ptr == nil {
		return nil
	}
	return *ptr
}

func nilOnEmptySlice[T any](s []T) *[]T {
	if len(s) == 0 {
		return nil
	}
	return &s
}
