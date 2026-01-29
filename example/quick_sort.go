package main

import (
	"sort"
	"sync"
)

const sortThreshold = 2048

type sortJob struct {
	lo, hi int
}

// ParallelQuickSort performs a parallel quicksort using the specified number of workers.
func ParallelQuickSort(arr []int, workers int) {
	if len(arr) <= 1 {
		return
	}

	jobs := make(chan sortJob, workers)
	var wg sync.WaitGroup

	spawnWorkers(workers, jobs, arr, &wg)

	wg.Add(1)
	jobs <- sortJob{0, len(arr) - 1}
	wg.Wait()

	close(jobs)
}

func spawnWorkers(count int, jobs chan sortJob, arr []int, wg *sync.WaitGroup) {
	for i := 0; i < count; i++ {
		go func() {
			for j := range jobs {
				quickSortPartition(arr, j.lo, j.hi, wg, jobs)
				wg.Done()
			}
		}()
	}
}

func quickSortPartition(arr []int, lo, hi int, wg *sync.WaitGroup, jobs chan<- sortJob) {
	if lo >= hi {
		return
	}

	// Use built-in sort for small partitions
	if hi-lo < sortThreshold {
		sort.Ints(arr[lo : hi+1])
		return
	}

	p := partition(arr, lo, hi)
	submitJob(arr, lo, p-1, wg, jobs)
	submitJob(arr, p+1, hi, wg, jobs)
}

func submitJob(arr []int, lo, hi int, wg *sync.WaitGroup, jobs chan<- sortJob) {
	if lo >= hi {
		return
	}

	wg.Add(1)
	select {
	case jobs <- sortJob{lo, hi}:
	default:
		// Fallback: execute synchronously if channel is full
		quickSortPartition(arr, lo, hi, wg, jobs)
		wg.Done()
	}
}

// partition uses median-of-three pivot selection for better performance.
func partition(arr []int, lo, hi int) int {
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
