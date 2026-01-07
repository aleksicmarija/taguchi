package main

import (
	"sort"
	"sync"
)

// ============================================================
// PARALLEL QUICK SORT
// ============================================================

func ParallelQuickSort(arr []int, workers int) {
	if len(arr) <= 1 {
		return
	}

	jobs := make(chan job, workers)
	var wg sync.WaitGroup

	worker := func() {
		for j := range jobs {
			quickSortJob(arr, j.lo, j.hi, &wg, jobs)
			wg.Done()
		}
	}

	for i := 0; i < workers; i++ {
		go worker()
	}

	wg.Add(1)
	jobs <- job{0, len(arr) - 1}
	wg.Wait()

	close(jobs)
}

// ============================================================
// QUICK SORT JOB
// ============================================================

func quickSortJob(arr []int, lo, hi int, wg *sync.WaitGroup, jobs chan<- job) {
	if lo >= hi {
		return

	}

	// Small partitions → Using built-in sort
	if hi-lo < 2048 {
		sort.Ints(arr[lo : hi+1])
		return

	}

	p := partitionMedian(arr, lo, hi)
	// Left partition
	submitJob(arr, lo, p-1, wg, jobs)
	// Right partition
	submitJob(arr, p+1, hi, wg, jobs)
}

// ============================================================
// JOB SUBMISSION
// ============================================================

func submitJob(arr []int, lo, hi int, wg *sync.WaitGroup, jobs chan<- job) {
	if lo >= hi {
		return
	}

	wg.Add(1)
	select {
	case jobs <- job{lo, hi}:
	// handled by worker
	default:
		// fallback: run synchronously
		quickSortJob(arr, lo, hi, wg, jobs)
		wg.Done()
	}
}

// ============================================================
// MEDIAN-OF-THREE PARTITION
// ============================================================

func partitionMedian(arr []int, lo, hi int) int {
	mid := lo + (hi-lo)/2

	// Order lo, mid, hi
	if arr[mid] < arr[lo] {
		arr[mid], arr[lo] = arr[lo], arr[mid]
	}
	if arr[hi] < arr[lo] {
		arr[hi], arr[lo] = arr[lo], arr[hi]
	}
	if arr[hi] < arr[mid] {
		arr[hi], arr[mid] = arr[mid], arr[hi]
	}

	// Move pivot to end
	arr[mid], arr[hi] = arr[hi], arr[mid]
	pivot := arr[hi]

	i := lo
	for j := lo; j < hi; j++ {
		if arr[j] <= pivot {
			arr[i], arr[j] = arr[j], arr[i]
			i++
		}
	}
	arr[i], arr[hi] = arr[hi], arr[i]
	return i
}
