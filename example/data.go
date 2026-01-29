package main

import "math/rand"

const randMax = 1_000_000

// generateData creates a slice of integers with the specified pattern.
func generateData(size int, pattern DataPattern) []int {
	data := make([]int, size)

	switch pattern {
	case Random:
		generateRandomPattern(data)
	case Sorted:
		generateSortedPattern(data)
	case ReverseSorted:
		generateReverseSortedPattern(data)
	case ManyDuplicates:
		generateDuplicatesPattern(data)
	case NearlySorted:
		generateNearlySortedPattern(data)
	}

	return data
}

func generateRandomPattern(data []int) {
	for i := range data {
		data[i] = rand.Intn(randMax)
	}
}

func generateSortedPattern(data []int) {
	for i := range data {
		data[i] = i
	}
}

func generateReverseSortedPattern(data []int) {
	for i := range data {
		data[i] = len(data) - i
	}
}

func generateDuplicatesPattern(data []int) {
	const maxValue = 100
	for i := range data {
		data[i] = rand.Intn(maxValue)
	}
}

func generateNearlySortedPattern(data []int) {
	for i := range data {
		data[i] = i
	}

	for i := 0; i < len(data)/10; i++ {
		a := rand.Intn(len(data))
		b := rand.Intn(len(data))
		data[a], data[b] = data[b], data[a]
	}
}

// isSorted checks if the array is sorted in ascending order.
func isSorted(arr []int) bool {
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[i-1] {
			return false
		}
	}
	return true
}
