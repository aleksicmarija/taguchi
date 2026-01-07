package main

import "sync"

// ============================================================
// PARALLEL RADIX SORT (LSD - Least Significant Digit first)
// ============================================================

func ParallelRadixSort(arr []int, workers int) {
	if len(arr) <= 1 {
		return
	}

	// Handle negative numbers by finding min and offsetting
	minVal := arr[0]
	for _, v := range arr {
		if v < minVal {
			minVal = v
		}
	}

	// Offset to make all values non-negative
	offset := 0
	if minVal < 0 {
		offset = -minVal
		for i := range arr {
			arr[i] += offset
		}
	}

	// Find max to determine number of digits
	maxVal := arr[0]
	for _, v := range arr {
		if v > maxVal {
			maxVal = v
		}
	}

	// Sort by each digit (base 256 for efficiency)
	const base = 256
	for exp := 1; maxVal/exp > 0; exp *= base {
		parallelCountingSort(arr, exp, base, workers)
	}

	// Restore original values
	if offset > 0 {
		for i := range arr {
			arr[i] -= offset
		}
	}
}

// ============================================================
// PARALLEL COUNTING SORT (for one digit position)
// ============================================================

func parallelCountingSort(arr []int, exp, base, workers int) {
	n := len(arr)
	output := make([]int, n)

	// Step 1: Parallel counting phase
	chunkSize := (n + workers - 1) / workers
	localCounts := make([][]int, workers)

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			counts := make([]int, base)
			start := workerID * chunkSize
			end := start + chunkSize
			if end > n {
				end = n
			}

			for i := start; i < end; i++ {
				digit := (arr[i] / exp) % base
				counts[digit]++
			}

			localCounts[workerID] = counts
		}(w)
	}
	wg.Wait()

	// Step 2: Merge local counts into global count
	globalCount := make([]int, base)
	for w := 0; w < workers; w++ {
		for digit := 0; digit < base; digit++ {
			globalCount[digit] += localCounts[w][digit]
		}
	}

	// Step 3: Convert to cumulative positions
	for i := 1; i < base; i++ {
		globalCount[i] += globalCount[i-1]
	}

	// Step 4: Place elements (must be sequential for stability)
	for i := n - 1; i >= 0; i-- {
		digit := (arr[i] / exp) % base
		globalCount[digit]--
		output[globalCount[digit]] = arr[i]
	}

	// Copy back
	copy(arr, output)
}
