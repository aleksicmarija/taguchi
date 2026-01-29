package main

import "sync"

const radixBase = 256

// ParallelRadixSort performs a parallel radix sort (LSD - Least Significant Digit first).
func ParallelRadixSort(arr []int, workers int) {
	if len(arr) <= 1 {
		return
	}

	offset := normalizeArray(arr)
	maxVal := findMax(arr)

	// Sort by each digit position
	for exp := 1; maxVal/exp > 0; exp *= radixBase {
		parallelCountingSort(arr, exp, workers)
	}

	// Restore original values
	if offset > 0 {
		for i := range arr {
			arr[i] -= offset
		}
	}
}

// normalizeArray handles negative numbers by offsetting all values to be non-negative.
// Returns the offset applied.
func normalizeArray(arr []int) int {
	minVal := arr[0]
	for _, v := range arr {
		if v < minVal {
			minVal = v
		}
	}

	offset := 0
	if minVal < 0 {
		offset = -minVal
		for i := range arr {
			arr[i] += offset
		}
	}

	return offset
}

// findMax returns the maximum value in the array.
func findMax(arr []int) int {
	maxVal := arr[0]
	for _, v := range arr {
		if v > maxVal {
			maxVal = v
		}
	}
	return maxVal
}

// parallelCountingSort performs a single-digit counting sort in parallel.
func parallelCountingSort(arr []int, exp, workers int) {
	n := len(arr)
	output := make([]int, n)

	// Phase 1: Parallel counting
	localCounts := countDigitsInParallel(arr, exp, workers)

	// Phase 2: Merge counts and build cumulative positions
	globalCount := mergeCountsAndCumulate(localCounts, workers)

	// Phase 3: Place elements (must be sequential for stability)
	placeElements(arr, output, globalCount, exp)

	// Copy back
	copy(arr, output)
}

func countDigitsInParallel(arr []int, exp, workers int) [][]int {
	n := len(arr)
	chunkSize := (n + workers - 1) / workers
	localCounts := make([][]int, workers)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			counts := make([]int, radixBase)
			start := workerID * chunkSize
			end := start + chunkSize
			if end > n {
				end = n
			}

			for i := start; i < end; i++ {
				digit := (arr[i] / exp) % radixBase
				counts[digit]++
			}

			localCounts[workerID] = counts
		}(w)
	}
	wg.Wait()

	return localCounts
}

func mergeCountsAndCumulate(localCounts [][]int, workers int) []int {
	globalCount := make([]int, radixBase)

	// Merge local counts
	for w := 0; w < workers; w++ {
		for digit := 0; digit < radixBase; digit++ {
			globalCount[digit] += localCounts[w][digit]
		}
	}

	// Convert to cumulative positions
	for i := 1; i < radixBase; i++ {
		globalCount[i] += globalCount[i-1]
	}

	return globalCount
}

func placeElements(arr, output []int, globalCount []int, exp int) {
	// Place elements in reverse order to maintain stability
	for i := len(arr) - 1; i >= 0; i-- {
		digit := (arr[i] / exp) % radixBase
		globalCount[digit]--
		output[globalCount[digit]] = arr[i]
	}
}
