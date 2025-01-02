package helpers

func MapSlice[T any, U any](slice []T, f func(T) U) []U {
	result := make([]U, len(slice))
	for i, v := range slice {
		result[i] = f(v)
	}
	return result
}

func FoldSlice[T any, U any, R any](slice []T, f func(T, R) R, initial R) R {
	result := initial
	for _, v := range slice {
		result = f(v, result)
	}
	return result
}
