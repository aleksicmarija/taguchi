package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/aleksicmarija/taguchi"
)

const dataSize = 2_000_000

// The noise factor. Every run is measured once per data pattern and the
// observations are pooled into one SNR, so the chosen settings have to be
// good across input orders rather than tuned to one of them.
var patterns = []DataPattern{Random, Sorted, ReverseSorted, ManyDuplicates, NearlySorted}

func main() {
	// Note: a more informative experiment would treat GOMAXPROCS as a factor
	// spanning larger core counts (for example 8, 16, 32) and add a factor
	// for a two-core configuration with fully isolated, dedicated CPUs, to
	// separate scalability from isolation effects. The example keeps things
	// simple and does not model that.
	workers := taguchi.NewFactor("MaxWorkers", 1, 20)
	algorithm := taguchi.NewFactor("Algorithm", QuickSort, RadixSort)
	procs := taguchi.NewFactor("GOMAXPROCS", 4, 8)

	// Column 2 of L8 carries the interaction of columns 0 and 1, so the third
	// factor goes on column 3 and any MaxWorkers x Algorithm interaction lands
	// in the error term instead of being mistaken for a GOMAXPROCS effect.
	design, err := taguchi.NewDesign(taguchi.L8.Select(0, 1, 3), workers, algorithm, procs)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(design)

	datasets := prepareDatasets(dataSize)
	exp := taguchi.NewExperiment(design)
	for _, run := range design.Runs() {
		runtime.GOMAXPROCS(procs.Of(run))
		for _, pattern := range patterns {
			data := make([]int, dataSize)
			copy(data, datasets[pattern])

			elapsed := sortWith(algorithm.Of(run), data, workers.Of(run))
			if !isSorted(data) {
				log.Fatalf("%v pattern=%s: output is not sorted", run, pattern)
			}
			fmt.Printf("%v pattern=%s: %v\n", run, pattern, elapsed)
			exp.Observe(run, float64(elapsed.Microseconds()))
		}
	}

	analysis, err := exp.Analyze(taguchi.SmallerTheBetter)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("\n", analysis)
	fmt.Printf("\nRecommendation: %s with MaxWorkers=%d on GOMAXPROCS=%d\n",
		algorithm.Best(analysis), workers.Best(analysis), procs.Best(analysis))
}

func prepareDatasets(size int) map[DataPattern][]int {
	datasets := make(map[DataPattern][]int, len(patterns))
	for _, p := range patterns {
		datasets[p] = generateData(size, p)
	}
	return datasets
}

func sortWith(alg SortAlgorithm, data []int, workers int) time.Duration {
	start := time.Now()
	switch alg {
	case QuickSort:
		ParallelQuickSort(data, workers)
	case RadixSort:
		ParallelRadixSort(data, workers)
	}
	return time.Since(start)
}
