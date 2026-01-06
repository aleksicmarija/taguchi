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
    "github.com/marijaaleksic/taguchi"
)

func main() {
    // Define control factors (what you can control)
    numThreads := taguchi.Factor{
        Name:   "NumThreads",
        Levels: []float64{1, 4, 8},
    }
    
    bufferSize := taguchi.Factor{
        Name:   "BufferSize",
        Levels: []float64{1024, 4096, 8192},
    }
    
    // Define noise factors (environmental conditions)
    cpuLoad := taguchi.NoiseFactor{
        Name:   "CPULoad",
        Levels: []float64{0.0, 0.5, 1.0},
    }
    
    // Create experiment with L9 orthogonal array
    exp, err := taguchi.NewExperiment(
        taguchi.SmallerTheBetter,  // Goal: minimize response time
        0.0,                        // Target (not used for SmallerTheBetter)
        0.05,                       // Pooling threshold for ANOVA
        []taguchi.Factor{numThreads, bufferSize},
        "L9",                       // Orthogonal array
        []taguchi.NoiseFactor{cpuLoad},
    )
    if err != nil {
        panic(err)
    }
    
    // Generate all trial combinations
    trials := exp.GenerateTrials()
    
    // Run experiments and collect observations
    for _, trial := range trials {
        numThreads := int(trial.Control["NumThreads"])
        bufferSize := int(trial.Control["BufferSize"])
        cpuLoad := trial.Noise["CPULoad"]
        
        // Run your experiment here
        observations := runYourExperiment(numThreads, bufferSize, cpuLoad)
        
        exp.AddResult(trial, observations)
    }
    
    // Analyze results
    results := exp.Analyze()
    
    // Display optimal settings
    fmt.Println("Optimal Factor Levels:")
    for factor, level := range results.OptimalLevels {
        fmt.Printf("  %s: %.2f\n", factor, level)
    }
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

Orthogonal arrays enable efficient experiment design by testing only a strategic subset of all possible combinations while maintaining statistical balance. The library includes standard arrays like L4, L8, L9, L16, L18, and L27.

## API Reference

### Types

#### `Factor`
Represents a controllable input variable.
```go
type Factor struct {
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

#### `AnalysisResult`
Complete analysis output.
```go
type AnalysisResult struct {
    OptimalLevels  map[string]float64      // Best factor levels
    SNR            map[string][]float64    // SNR for each level
    MainEffects    map[string][]float64    // Average SNR per level
    Contributions  map[string]float64      // Factor importance (%)
    ANOVA          ANOVAResult             // Detailed statistics
}
```

### Methods

#### `NewExperiment`
```go
func NewExperiment(
    goal OptimizationGoal,
    target float64,
    poolingThreshold float64,
    controlFactors []Factor,
    arrayName string,
    noiseFactors []NoiseFactor,
) (*Experiment, error)
```
Creates a new Taguchi experiment with specified parameters.

#### `GenerateTrials`
```go
func (e *Experiment) GenerateTrials() []Trial
```
Generates all trial combinations from the orthogonal array and noise factors.

#### `AddResult`
```go
func (e *Experiment) AddResult(trial Trial, observations []float64)
```
Records experimental observations for a trial.

#### `Analyze`
```go
func (e *Experiment) Analyze() AnalysisResult
```
Performs complete statistical analysis including ANOVA and optimal level determination.

## Example: Parallel Sorting Optimization

See `examples/main.go` for a complete example that optimizes parallel sorting algorithms by varying:

- **Control Factors**: Number of goroutines, sorting algorithm (QuickSort, MergeSort, BitonicSort)
- **Noise Factors**: CPU load levels
- **Goal**: Minimize sorting time

The example demonstrates:
- Setting up a multi-factor experiment
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