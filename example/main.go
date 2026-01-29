package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/marijaaleksic/taguchi"
)

const dataSize = 2_000_000

func main() {
	exp, err := createExperiment()
	if err != nil {
		log.Fatal(err)
	}

	datasets := prepareDatasets(dataSize)
	runExperiment(exp, datasets)

	results := exp.Analyze()
	taguchi.PrintAnalysisReport(results)
}

func createExperiment() (*taguchi.Experiment, error) {
	factors := []taguchi.Factor{
		{Name: "MaxWorkers", Levels: []float64{1, 20}},
		{Name: "Algorithm", Levels: []float64{0, 1}},
		{Name: "GOMAXPROCS", Levels: []float64{4, 8}},
	}

	noise := []taguchi.NoiseFactor{
		{Name: "DataPattern", Levels: []float64{0, 1, 2, 3, 4}},
	}

	return taguchi.NewExperiment(
		taguchi.SmallerTheBetter,
		0,
		0.05,
		factors,
		"L4",
		noise,
	)
}

func prepareDatasets(size int) map[DataPattern][]int {
	patterns := []DataPattern{Random, Sorted, ReverseSorted, ManyDuplicates, NearlySorted}
	datasets := make(map[DataPattern][]int, len(patterns))

	for _, p := range patterns {
		datasets[p] = generateData(size, p)
	}

	return datasets
}

func runExperiment(exp *taguchi.Experiment, datasets map[DataPattern][]int) {
	for _, trial := range exp.GenerateTrials() {
		tc := trialConfig{trial: trial, datasets: datasets}
		runTrial(exp, tc)
	}
}

type trialConfig struct {
	trial    taguchi.Trial
	datasets map[DataPattern][]int
}

func runTrial(exp *taguchi.Experiment, tc trialConfig) {
	runtime.GOMAXPROCS(int(tc.trial.Control["GOMAXPROCS"]))

	workers := int(tc.trial.Control["MaxWorkers"])
	alg := SortAlgorithm(tc.trial.Control["Algorithm"])
	pattern := DataPattern(tc.trial.Noise["DataPattern"])

	data := make([]int, dataSize)
	copy(data, tc.datasets[pattern])

	printTrialStart(tc.trial, alg, workers, pattern)

	dur := executeSortAlgorithm(alg, data, workers)

	if !isSorted(data) {
		panic("sorting failed")
	}

	exp.AddResult(tc.trial, []float64{float64(dur.Microseconds())})
	printTrialResult(tc.trial, alg, workers, pattern, dur)
}

func executeSortAlgorithm(alg SortAlgorithm, data []int, workers int) time.Duration {
	start := time.Now()

	switch alg {
	case QuickSort:
		ParallelQuickSort(data, workers)
	case RadixSort:
		ParallelRadixSort(data, workers)
	}

	return time.Since(start)
}

func printTrialStart(trial taguchi.Trial, alg SortAlgorithm, workers int, pattern DataPattern) {
	fmt.Printf("Trial %d: %s | Workers=%d | GOMAXPROCS=%d | Pattern=%s\n",
		trial.ID, alg, workers, int(trial.Control["GOMAXPROCS"]), pattern)
}

func printTrialResult(trial taguchi.Trial, alg SortAlgorithm, workers int, pattern DataPattern, dur time.Duration) {
	fmt.Printf("  Result: %v\n\n", dur)
}
