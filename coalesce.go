package restapi

// coalesce returns the first non-zero value from a list of values.
func coalesce[T comparable](values ...T) T {
	z := *new(T)
	for _, v := range values {
		if v != z {
			return v
		}
	}
	return z
}

// ifNil returns the value of a pointer if it is not nil, or a specified
// value if the pointer is nil.
func ifNil[T any](ptr *T, value T) T {
	if ptr != nil {
		return *ptr
	}
	return value
}

// ifNotNil returns the value referenced by a pointer if it is not nil,
// or the zero value of the type referenced by the pointer if the pointer
// is nil.
func ifNotNil[T any](ptr *T) T {
	if ptr != nil {
		return *ptr
	}
	return *new(T)
}
