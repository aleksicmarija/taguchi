package main

import (
	"slices"
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

	jobs := make(chan sortJob, workers*128)
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
		slices.Sort(arr[lo : hi+1])
		return
	}

	lt, gt := partition3Way(arr, lo, hi)
	submitJob(arr, lo, lt-1, wg, jobs)
	submitJob(arr, gt+1, hi, wg, jobs)
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

// partition3Way uses the Dutch National Flag algorithm with median-of-three pivot selection.
// Returns (lt, gt) such that arr[lo..lt-1] < pivot, arr[lt..gt] == pivot, arr[gt+1..hi] > pivot.
func partition3Way(arr []int, lo, hi int) (int, int) {
	mid := lo + (hi-lo)/2

	// Median-of-three pivot selection
	if arr[mid] < arr[lo] {
		arr[mid], arr[lo] = arr[lo], arr[mid]
	}
	if arr[hi] < arr[lo] {
		arr[hi], arr[lo] = arr[lo], arr[hi]
	}
	if arr[hi] < arr[mid] {
		arr[hi], arr[mid] = arr[mid], arr[hi]
	}
	pivot := arr[mid]

	lt, i, gt := lo, lo, hi
	for i <= gt {
		if arr[i] < pivot {
			arr[lt], arr[i] = arr[i], arr[lt]
			lt++
			i++
		} else if arr[i] > pivot {
			arr[i], arr[gt] = arr[gt], arr[i]
			gt--
		} else {
			i++
		}
	}
	return lt, gt
}
