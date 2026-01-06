package main

import (
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"github.com/marijaaleksic/taguchi"
)

type SortAlgorithm int

const (
	QuickSort SortAlgorithm = iota
	MergeSort
	BitonicSort
)

func (s SortAlgorithm) String() string {
	switch s {
	case QuickSort:
		return "QuickSort"
	case MergeSort:
		return "MergeSort"
	case BitonicSort:
		return "BitonicSort"
	default:
		return "Unknown"
	}
}

// ==========================
// Parallel Quick Sort (In-place)
// ==========================

func parallelQuickSort(arr []int, numGoroutines int) {
	if len(arr) <= 1 {
		return
	}

	var wg sync.WaitGroup
	recursiveQuickSort(arr, numGoroutines, &wg)
	wg.Wait()
}

func recursiveQuickSort(arr []int, depth int, wg *sync.WaitGroup) {
	if len(arr) <= 1 {
		return
	}

	if len(arr) < 1000 || depth <= 1 {
		sort.Ints(arr)
		return
	}

	pivotIdx := partition(arr)

	if depth > 1 {
		wg.Add(2)
		newDepth := depth / 2

		go func() {
			defer wg.Done()
			recursiveQuickSort(arr[:pivotIdx], newDepth, wg)
		}()

		go func() {
			defer wg.Done()
			recursiveQuickSort(arr[pivotIdx+1:], newDepth, wg)
		}()
	} else {
		recursiveQuickSort(arr[:pivotIdx], depth, wg)
		recursiveQuickSort(arr[pivotIdx+1:], depth, wg)
	}
}

func partition(arr []int) int {
	pivot := arr[len(arr)-1]
	i := -1

	for j := 0; j < len(arr)-1; j++ {
		if arr[j] <= pivot {
			i++
			arr[i], arr[j] = arr[j], arr[i]
		}
	}
	arr[i+1], arr[len(arr)-1] = arr[len(arr)-1], arr[i+1]
	return i + 1
}

// ==========================
// Parallel Merge Sort (In-place)
// ==========================

func parallelMergeSort(arr []int, numGoroutines int) {
	if len(arr) <= 1 {
		return
	}
	temp := make([]int, len(arr))
	recursiveMergeSortInPlace(arr, temp, 0, len(arr)-1, numGoroutines)
}

func recursiveMergeSortInPlace(arr, temp []int, left, right, depth int) {
	if left >= right {
		return
	}

	if right-left < 1000 || depth <= 1 {
		sort.Ints(arr[left : right+1])
		return
	}

	mid := (left + right) / 2

	if depth > 1 {
		var wg sync.WaitGroup
		wg.Add(2)
		newDepth := depth / 2

		go func() {
			defer wg.Done()
			recursiveMergeSortInPlace(arr, temp, left, mid, newDepth)
		}()

		go func() {
			defer wg.Done()
			recursiveMergeSortInPlace(arr, temp, mid+1, right, newDepth)
		}()

		wg.Wait()
	} else {
		recursiveMergeSortInPlace(arr, temp, left, mid, depth)
		recursiveMergeSortInPlace(arr, temp, mid+1, right, depth)
	}

	mergeInPlace(arr, temp, left, mid, right)
}

func mergeInPlace(arr, temp []int, left, mid, right int) {
	i, j, k := left, mid+1, left

	for i <= mid && j <= right {
		if arr[i] <= arr[j] {
			temp[k] = arr[i]
			i++
		} else {
			temp[k] = arr[j]
			j++
		}
		k++
	}

	for i <= mid {
		temp[k] = arr[i]
		i++
		k++
	}

	for j <= right {
		temp[k] = arr[j]
		j++
		k++
	}

	copy(arr[left:right+1], temp[left:right+1])
}

// ==========================
// Parallel Bitonic Sort (Fixed)
// ==========================

func parallelBitonicSort(arr []int, numGoroutines int) {
	n := len(arr)
	origLen := n

	// Pad to power of 2 if needed
	if n&(n-1) != 0 {
		nextPow2 := 1
		for nextPow2 < n {
			nextPow2 <<= 1
		}
		// Pad with max int so they sort to the end
		for len(arr) < nextPow2 {
			arr = append(arr, int(^uint(0)>>1)) // Max int
		}
		n = nextPow2
	}

	bitonicSortRecursive(arr, 0, n, true, numGoroutines)

	// Remove padding if we added any
	arr = arr[:origLen]
}

func bitonicSortRecursive(arr []int, low, cnt int, dir bool, depth int) {
	if cnt > 1 {
		k := cnt / 2

		if depth > 1 && cnt > 1000 {
			var wg sync.WaitGroup
			wg.Add(2)
			newDepth := depth / 2

			go func() {
				defer wg.Done()
				bitonicSortRecursive(arr, low, k, true, newDepth)
			}()

			go func() {
				defer wg.Done()
				bitonicSortRecursive(arr, low+k, k, false, newDepth)
			}()

			wg.Wait()
		} else {
			bitonicSortRecursive(arr, low, k, true, depth)
			bitonicSortRecursive(arr, low+k, k, false, depth)
		}

		bitonicMerge(arr, low, cnt, dir, depth)
	}
}

func bitonicMerge(arr []int, low, cnt int, dir bool, depth int) {
	if cnt > 1 {
		k := cnt / 2
		for i := low; i < low+k; i++ {
			compareAndSwap(arr, i, i+k, dir)
		}

		if depth > 1 && cnt > 1000 {
			var wg sync.WaitGroup
			wg.Add(2)
			newDepth := depth / 2

			go func() {
				defer wg.Done()
				bitonicMerge(arr, low, k, dir, newDepth)
			}()

			go func() {
				defer wg.Done()
				bitonicMerge(arr, low+k, k, dir, newDepth)
			}()

			wg.Wait()
		} else {
			bitonicMerge(arr, low, k, dir, depth)
			bitonicMerge(arr, low+k, k, dir, depth)
		}
	}
}

func compareAndSwap(arr []int, i, j int, dir bool) {
	if dir == (arr[i] > arr[j]) {
		arr[i], arr[j] = arr[j], arr[i]
	}
}

// ==========================
// CPU Load Generator (Noise)
// ==========================

func generateCPULoad(intensity float64, duration time.Duration, done chan bool) {
	numWorkers := int(intensity * 10)

	var wg sync.WaitGroup
	stopChan := make(chan bool, numWorkers)

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			fibonacci := func(n int) int {
				if n <= 1 {
					return n
				}
				a, b := 0, 1
				for i := 2; i <= n; i++ {
					a, b = b, a+b
				}
				return b
			}

			for {
				select {
				case <-stopChan:
					return
				default:
					sum := 0
					for j := 0; j < 100; j++ {
						sum += fibonacci(30 + rand.Intn(10))
					}
					_ = sum
				}
			}
		}()
	}

	select {
	case <-done:
		close(stopChan)
	case <-time.After(duration):
		close(stopChan)
	}

	wg.Wait()
}

// ==========================
// Run Single Trial
// ==========================

func runSortingTrial(algorithm SortAlgorithm, numGoroutines int, cpuLoad float64, originalData []int) time.Duration {
	// Create a fresh copy of the data for this trial
	data := make([]int, len(originalData))
	copy(data, originalData)

	// Start CPU load
	done := make(chan bool)
	go generateCPULoad(cpuLoad, 10*time.Second, done)

	// Let CPU load stabilize
	time.Sleep(100 * time.Millisecond)

	// Measure sorting time
	start := time.Now()

	switch algorithm {
	case QuickSort:
		parallelQuickSort(data, numGoroutines)
	case MergeSort:
		parallelMergeSort(data, numGoroutines)
	case BitonicSort:
		parallelBitonicSort(data, numGoroutines)
	}

	elapsed := time.Since(start)

	// Stop CPU load
	close(done)

	return elapsed
}

// ==========================
// Main Experiment
// ==========================

func main() {
	rand.Seed(time.Now().UnixNano())

	numGoroutinesFactor := taguchi.Factor{
		Name:   "NumGoroutines",
		Levels: []float64{1, 4, 8},
	}

	algorithmFactor := taguchi.Factor{
		Name:   "Algorithm",
		Levels: []float64{0, 1, 2},
	}

	cpuLoadNoise := taguchi.NoiseFactor{
		Name:   "CPULoad",
		Levels: []float64{0.0, 0.3, 0.7},
	}

	exp, err := taguchi.NewExperiment(
		taguchi.SmallerTheBetter,
		0.0,
		0.05,
		[]taguchi.Factor{numGoroutinesFactor, algorithmFactor},
		"L9",
		[]taguchi.NoiseFactor{cpuLoadNoise},
	)

	if err != nil {
		panic(err)
	}

	trials := exp.GenerateTrials()

	fmt.Printf("Running %d trials...\n\n", len(trials))

	dataSize := 500000
	repetitions := 3

	// Generate test data ONCE - all algorithms will use the same data
	fmt.Println("Generating test data...")
	testData := make([]int, dataSize)
	for i := range testData {
		testData[i] = rand.Intn(100000)
	}

	for _, trial := range trials {
		numGoroutines := int(trial.Control["NumGoroutines"])
		algorithmID := int(trial.Control["Algorithm"])
		cpuLoad := trial.Noise["CPULoad"]

		algorithm := SortAlgorithm(algorithmID)

		fmt.Printf("Trial %d: Algorithm=%s, Goroutines=%d, CPULoad=%.1f%%\n",
			trial.ID, algorithm, numGoroutines, cpuLoad*100)

		observations := make([]float64, repetitions)
		for rep := 0; rep < repetitions; rep++ {
			duration := runSortingTrial(algorithm, numGoroutines, cpuLoad, testData)
			observations[rep] = float64(duration.Microseconds())
			fmt.Printf("  Rep %d: %v\n", rep+1, duration)
		}

		exp.AddResult(trial, observations)

		// pause between trials to let system stabilize
		time.Sleep(500 * time.Millisecond)
	}

	fmt.Println("\n" + "============================================================")
	fmt.Println("ANALYSIS RESULTS")
	fmt.Println("============================================================")

	results := exp.Analyze()

	fmt.Println("Optimal Factor Levels:")
	fmt.Println("----------------------------------------")
	for factor, level := range results.OptimalLevels {
		if factor == "NumGoroutines" {
			fmt.Printf("  %s: %d goroutines\n", factor, int(level))
		} else if factor == "Algorithm" {
			fmt.Printf("  %s: %s\n", factor, SortAlgorithm(int(level)))
		} else {
			fmt.Printf("  %s: %.2f\n", factor, level)
		}
	}

	fmt.Println("\nMain Effects (Average SNR per level):")
	fmt.Println("----------------------------------------")
	for factor, effects := range results.MainEffects {
		fmt.Printf("  %s:\n", factor)
		for i, effect := range effects {
			if factor == "NumGoroutines" {
				goroutines := int(exp.ControlFactors[0].Levels[i])
				if exp.ControlFactors[0].Name != "NumGoroutines" {
					goroutines = int(exp.ControlFactors[1].Levels[i])
				}
				fmt.Printf("    Level %d (%d goroutines): %.2f\n", i+1, goroutines, effect)
			} else if factor == "Algorithm" {
				fmt.Printf("    Level %d (%s): %.2f\n", i+1, SortAlgorithm(i), effect)
			} else {
				fmt.Printf("    Level %d: %.2f\n", i+1, effect)
			}
		}
	}

	fmt.Println("\nFactor Contributions (%):")
	fmt.Println("----------------------------------------")
	for factor, contrib := range results.Contributions {
		fmt.Printf("  %s: %.2f%%\n", factor, contrib)
	}

	fmt.Println("\nANOVA Results:")
	fmt.Println("----------------------------------------")
	fmt.Printf("  %-20s %12s %8s %12s %12s\n", "Factor", "SS", "DF", "MS", "F-ratio")
	fmt.Println("  " + "------------------------------------------------------------")

	for factor := range results.ANOVA.FactorSS {
		fmt.Printf("  %-20s %12.2f %8d %12.2f %12.2f\n",
			factor,
			results.ANOVA.FactorSS[factor],
			results.ANOVA.FactorDF[factor],
			results.ANOVA.FactorMS[factor],
			results.ANOVA.FactorF[factor])
	}

	fmt.Printf("  %-20s %12.2f %8d %12.2f\n",
		"Error",
		results.ANOVA.ErrorSS,
		results.ANOVA.ErrorDF,
		results.ANOVA.ErrorMS)

	fmt.Println("\n" + "============================================================")
	fmt.Println("Experiment Complete!")
	fmt.Println("============================================================")
}
