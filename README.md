# Taguchi Method Library for Go

Go library for conducting Taguchi Method experiments (Design of Experiments) to optimize system parameters through statistical analysis.

## Overview

The Taguchi Method is a statistical technique for improving product quality and process optimization. This library provides a complete implementation for designing experiments, collecting data, and analyzing results using orthogonal arrays, Signal-to-Noise (SNR) ratios, and ANOVA.

## Features

- **Multiple Optimization Goals**: Support for Smaller-the-Better, Larger-the-Better, and Nominal-the-Best quality characteristics
- **Orthogonal Array Support**: Built-in standard arrays (L4, L8, L9, etc.) for experiment design
- **Noise Factor Modeling**: Parameter design with controllable and uncontrollable factors
- **Analysis**: ANOVA calculations including F-ratios, contributions, and optimal levels
- **Trial Generation**: Automatic generation of all experimental combinations
- **Main Effects Analysis**: Identification of factor impacts on performance

## Installation

```bash
go get github.com/marijaaleksic/taguchi
```

## Quick Start

```go
package main

import (
	"fmt"
	"log"
	"runtime"
	"time"

	"github.com/marijaaleksic/taguchi"
)

type ExperimentFactors struct {
	MaxWorkers []float64
	Algorithm  []float64
	GOMAXPROCS []float64
}

func main() {
	// Define control factors as a struct with []float64 fields
	factors := ExperimentFactors{
		MaxWorkers: []float64{1, 20},
		Algorithm:  []float64{0, 1},
		GOMAXPROCS: []float64{4, 8},
	}

	// Define noise factors (environmental conditions)
	noise := []taguchi.NoiseFactor{
		{Name: "DataPattern", Levels: []float64{0, 1, 2, 3, 4}},
	}

	// Create experiment with L4 orthogonal array
	exp, err := taguchi.NewExperiment[ExperimentFactors, ExperimentFactors](
		&taguchi.SmallerTheBetter{},
		factors,
		"L4",
		noise,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Generate all trial combinations
	trials := exp.GenerateTrials()

	// Run experiments and collect observations
	for _, trial := range trials {
		params := exp.Params(trial)  // Convert trial to factors struct
		
		// Access factor values
		workers := int(params.MaxWorkers)
		alg := int(params.Algorithm)
		gomaxprocs := int(params.GOMAXPROCS)
		pattern := int(trial.Noise["DataPattern"])

		// Run your experiment
		runtime.GOMAXPROCS(gomaxprocs)
		duration := runYourExperiment(workers, alg, pattern)
		
		// Record observations
		exp.AddResult(trial, []float64{float64(duration.Microseconds())})
	}

	// Analyze results
	results := exp.Analyze()
	taguchi.PrintAnalysisReport(results)
}

func runYourExperiment(workers, alg, pattern int) time.Duration {
	start := time.Now()
	// Your experiment logic here
	return time.Since(start)
}
```

## Core Concepts

### Optimization Goals

The library supports three quality characteristic types:

- **SmallerTheBetter**: Minimize the response (e.g., defects, cost, time)
- **LargerTheBetter**: Maximize the response (e.g., strength, yield, throughput)
- **NominalTheBest**: Hit a specific target value with minimal variation

### Signal-to-Noise Ratio (SNR)

SNR quantifies the robustness of a design:

- **Smaller-the-Better**: SNR = -10 × log₁₀(mean(y²))
- **Larger-the-Better**: SNR = -10 × log₁₀(mean(1/y²))
- **Nominal-the-Best**: SNR = -10 × log₁₀(mean((y - target)²))

Higher SNR values indicate better performance with less sensitivity to noise.

### Orthogonal Arrays

Orthogonal arrays enable efficient experiment design by testing only a strategic subset of all possible combinations while maintaining statistical balance. The library includes standard arrays like L4, L8, L9, L16, and L18.

## API Reference

## API Reference

### Types

#### `ControlFactor`
Represents a controllable input variable (when using manual factor construction).
```go
type ControlFactor struct {
    Name   string      // Factor identifier
    Levels []float64   // Possible values
}
```

#### `NoiseFactor`
Represents an uncontrollable environmental variable.
```go
type NoiseFactor struct {
    Name   string      // Noise factor identifier
    Levels []float64   // Environmental conditions
}
```

#### `Trial`
A single experimental configuration.
```go
type Trial struct {
    ID      int
    Control map[string]float64  // Factor settings
    Noise   map[string]float64  // Environmental conditions
}
```

#### `TrialResult`
Records observations from a completed trial.
```go
type TrialResult struct {
    Trial        Trial           // The experimental configuration
    Observations []float64       // Measured results
}
```

#### `AnalysisResult`
Complete analysis output.
```go
type AnalysisResult struct {
    OptimalLevels map[string]float64      // Best factor levels
    SNR           map[string][]float64    // SNR for each level
    MainEffects   map[string][]float64    // Average SNR per level
    Contributions map[string]float64      // Factor importance (%)
    ANOVA         ANOVAResult             // Detailed statistics
}
```

#### `ANOVAResult`
Detailed ANOVA statistics.
```go
type ANOVAResult struct {
    FactorSS      map[string]float64  // Sum of squares per factor
    FactorDF      map[string]int      // Degrees of freedom per factor
    FactorMS      map[string]float64  // Mean square per factor
    FactorF       map[string]float64  // F-ratio per factor
    ErrorSS       float64             // Sum of squares for error
    ErrorDF       int                 // Degrees of freedom for error
    ErrorMS       float64             // Mean square error
    PooledFactors []string            // Factors pooled during analysis
}
```

#### `OptimizationGoal`
Interface for quality characteristics.
```go
type OptimizationGoal interface {
    CalculateSNR(observations []float64) float64
    String() string
}

// Built-in implementations:
type SmallerTheBetter struct{}      // Minimize response
type LargerTheBetter struct{}       // Maximize response
type NominalTheBest struct {
    Target float64                  // Achieve target value
}
```

### Methods

#### `NewExperiment` (Generic with Struct Factors)
```go
func NewExperiment[F any, P any](
    goal OptimizationGoal,
    factors F,
    arrayName string,
    noiseFactors []NoiseFactor,
) (*Experiment[P], error)
```
Creates a new Taguchi experiment. `F` is the factors struct type (inferred from the factors argument), `P` is the params struct type for converting trials to factor values.

#### `NewExperimentUsingArray` (Generic with Custom Array)
```go
func NewExperimentUsingArray[F any, P any](
    goal OptimizationGoal,
    factors F,
    orthogonalArray [][]int,
    noiseFactors []NoiseFactor,
) (*Experiment[P], error)
```
Creates a new Taguchi experiment with a user-provided custom orthogonal array.

#### `NewExperimentFromFactors` (Manual Factor Construction)
```go
func NewExperimentFromFactors(
    goal OptimizationGoal,
    controlFactors []ControlFactor,
    arrayName string,
    noiseFactors []NoiseFactor,
) (*Experiment[struct{}], error)
```
Creates a new Taguchi experiment from pre-built ControlFactor slices instead of struct fields.

#### `NewExperimentFromFactorsUsingArray` (Manual with Custom Array)
```go
func NewExperimentFromFactorsUsingArray(
    goal OptimizationGoal,
    controlFactors []ControlFactor,
    orthogonalArray [][]int,
    noiseFactors []NoiseFactor,
) (*Experiment[struct{}], error)
```
Creates a new Taguchi experiment from pre-built ControlFactor slices with a custom orthogonal array.

#### `Params`
```go
func (e *Experiment[P]) Params(trial Trial) P
```
Converts a Trial's Control map into a value of type P. Exported float64 fields are populated from control factor values.

#### `GenerateTrials`
```go
func (e *Experiment[P]) GenerateTrials() []Trial
```
Generates all trial combinations from the orthogonal array and noise factors.

#### `AddResult`
```go
func (e *Experiment[P]) AddResult(trial Trial, observations []float64)
```
Records experimental observations for a trial.

#### `Analyze`
```go
func (e *Experiment[P]) Analyze() AnalysisResult
```
Performs complete statistical analysis including ANOVA and optimal level determination.

#### `PrintAnalysisReport`
```go
func PrintAnalysisReport(result AnalysisResult)
```
Prints a formatted analysis report to stdout.

## Example: Parallel Sorting Optimization

See `example/main.go` for a complete example that optimizes parallel sorting algorithms by varying:

- **Control Factors**: Number of workers, sorting algorithm (QuickSort, RadixSort), GOMAXPROCS
- **Noise Factors**: Data patterns (random, sorted, reverse sorted, many duplicates, nearly sorted)
- **Goal**: Minimize sorting time

The example demonstrates:
- Defining control factors as a struct with `[]float64` fields
- Running trials with environmental noise
- Analyzing results to find optimal configurations
- Interpreting ANOVA and contribution percentages

## Understanding the Output

### Main Effects
Shows the average SNR for each factor level. Higher values indicate better performance.

### Factor Contributions
Percentage contribution of each factor to total variation. Higher percentages mean the factor has more impact on performance.

### ANOVA Results
- **SS (Sum of Squares)**: Variation attributed to each factor
- **DF (Degrees of Freedom)**: Number of independent factor levels minus one
- **MS (Mean Square)**: SS divided by DF
- **F-ratio**: Factor significance (higher values indicate more significant factors)

### Optimal Levels
The factor settings that maximize SNR (i.e., best performance with least variation).

## Best Practices

1. **Choose Appropriate Arrays**: Select an orthogonal array that can accommodate all your factors
2. **Multiple Observations**: Run multiple repetitions per trial to capture variation
3. **Noise Factors**: Include realistic environmental conditions that affect your system
4. **Goal Selection**: Choose the optimization goal that matches your quality characteristic
5. **Analyze Contributions**: Focus optimization efforts on high-contribution factors

## License

This is free and unencumbered software released into the public domain.

Anyone is free to copy, modify, publish, use, compile, sell, or distribute this software, for any purpose, commercial or non-commercial, without any conditions.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.