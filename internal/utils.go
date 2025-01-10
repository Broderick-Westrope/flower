package internal

func ToPointer[T any](v T) *T {
	return &v
}

func AreMutuallyExclusive(bools ...bool) bool {
	count := 0
	for _, b := range bools {
		if b {
			count++
		}
		if count > 1 {
			return false
		}
	}
	return count <= 1
}
