package main

// ReversePaginate returns elements from a slice in reverse chronological order.
// Elements are assumed to be ordered from oldest to newest in the input slice.
// Returns the specified page of results with newest items first.
func ReversePaginate[T any](elements []T, pageNumber, countPerPage int) []T {
	if pageNumber < 1 || countPerPage < 1 || len(elements) == 0 {
		return []T{}
	}

	totalElements := len(elements)
	startIndex := totalElements - (pageNumber-1)*countPerPage - 1
	endIndex := totalElements - pageNumber*countPerPage

	if startIndex < 0 {
		return []T{}
	}

	if endIndex < -1 {
		endIndex = -1
	}

	result := make([]T, 0, startIndex-endIndex)
	for i := startIndex; i > endIndex; i-- {
		result = append(result, elements[i])
	}

	return result
}
