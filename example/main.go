package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/marijaaleksic/taguchi"
)

// ==========================
// Algorithms
// ==========================

type SortAlgorithm int

const (
	QuickSort SortAlgorithm = iota
	RadixSort
)

func (s SortAlgorithm) String() string {
	switch s {
	case QuickSort:
		return "QuickSort"
	case RadixSort:
		return "RadixSort"
	default:
		return "Unknown"
	}
}

// ==========================
// Data Patterns (Noise)
// ==========================

type DataPattern int

const (
	Random DataPattern = iota
	Sorted
	ReverseSorted
	ManyDuplicates
	NearlySorted
)

func (d DataPattern) String() string {
	switch d {
	case Random:
		return "Random"
	case Sorted:
		return "Sorted"
	case ReverseSorted:
		return "ReverseSorted"
	case ManyDuplicates:
		return "ManyDuplicates"
	case NearlySorted:
		return "NearlySorted"
	default:
		return "Unknown"
	}
}

// ============================================================
// JOB TYPE
// ============================================================

type job struct {
	lo, hi int
}

// ============================================================
// DATA GENERATION & VERIFICATION
// ============================================================

func generateData(size int, pattern DataPattern) []int {
	data := make([]int, size)
	switch pattern {
	case Random:
		for i := range data {
			data[i] = rand.Intn(1_000_000)
		}
	case Sorted:
		for i := range data {
			data[i] = i
		}
	case ReverseSorted:
		for i := range data {
			data[i] = size - i
		}
	case ManyDuplicates:
		for i := range data {
			data[i] = rand.Intn(100)
		}
	case NearlySorted:
		for i := range data {
			data[i] = i
		}
		for i := 0; i < size/10; i++ {
			a := rand.Intn(size)
			b := rand.Intn(size)
			data[a], data[b] = data[b], data[a]
		}
	}
	return data
}

func isSorted(arr []int) bool {
	for i := 1; i < len(arr); i++ {
		if arr[i] < arr[i-1] {
			return false
		}
	}
	return true
}

// ============================================================
// FACTOR & PARAM STRUCTS (typed config for Taguchi factors)
// ============================================================

type SortFactors struct {
	MaxWorkers []float64
	Algorithm  []float64
}

type SortParams struct {
	MaxWorkers float64
	Algorithm  float64
}

// ============================================================
// MAIN EXPERIMENT
// ============================================================

func main() {
	noise := taguchi.NoiseFactor{
		Name:   "DataPattern",
		Levels: []float64{0, 1, 2, 3, 4},
	}

	exp, err := taguchi.NewExperiment[SortFactors, SortParams](
		taguchi.SmallerTheBetter{},
		SortFactors{
			MaxWorkers: []float64{6, 9, 15, 20},
			Algorithm:  []float64{0, 1},
		},
		taguchi.L8,
		[]taguchi.NoiseFactor{noise},
	)
	if err != nil {
		panic(err)
	}

	dataSize := 2_000_000
	patterns := []DataPattern{Random, Sorted, ReverseSorted, ManyDuplicates, NearlySorted}
	datasets := map[DataPattern][]int{}
	for _, p := range patterns {
		datasets[p] = generateData(dataSize, p)
	}

	for _, trial := range exp.GenerateTrials() {
		params := exp.Params(trial)
		workers := int(params.MaxWorkers)
		alg := SortAlgorithm(params.Algorithm)
		pattern := DataPattern(trial.Noise["DataPattern"])

		data := make([]int, dataSize)
		copy(data, datasets[pattern])

		start := time.Now()
		fmt.Println("Running Trial:", trial.ID, "Algorithm:", alg, "Workers:", workers, "Pattern:", pattern)
		if alg == QuickSort {
			ParallelQuickSort(data, workers)
		} else {
			ParallelRadixSort(data, workers)
		}
		dur := time.Since(start)

		if !isSorted(data) {
			panic("sorting failed")
		}

		exp.AddResult(trial, []float64{float64(dur.Microseconds())})

		fmt.Printf("Trial %d | %s | workers=%d | %s | %v\n",
			trial.ID, alg, workers, pattern, dur)
	}

	results := exp.Analyze()
	taguchi.PrintAnalysisReport(results)
}
