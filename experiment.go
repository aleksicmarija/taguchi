package taguchi

import (
	"fmt"
)

// NewExperiment initializes a new Taguchi experiment with control factors, noise factors, goal, and target.
// arrayName selects a standard orthogonal array (e.g., "L4", "L8") to generate the trial layout.
// Returns an error if the array does not exist or cannot accommodate the number of control factors.
func NewExperiment(goal OptimizationGoal, target float64, poolingThreshold float64, controlFactors []Factor, arrayName string, noiseFactors []NoiseFactor) (*Experiment, error) {
	oa, ok := StandardArrays[arrayName]
	if !ok {
		return nil, fmt.Errorf("orthogonal array %s not defined", arrayName)
	}
	if len(controlFactors) > len(oa[0]) {
		return nil, fmt.Errorf("orthogonal array %s cannot accommodate %d factors", arrayName, len(controlFactors))
	}
	return &Experiment{
		ControlFactors:   controlFactors,
		NoiseFactors:     noiseFactors,
		Goal:             goal,
		Target:           target,
		OrthogonalArray:  oa,
		PoolingThreshold: poolingThreshold,
	}, nil
}

// AddResult records the observations from a completed trial into the experiment's results.
// trial: The specific trial that was executed.
// observations: Slice of measured outcomes for this trial.
func (e *Experiment) AddResult(trial Trial, observations []float64) {
	e.Results = append(e.Results, TrialResult{
		Trial:        trial,
		Observations: observations,
	})
}

// Analyze performs a full Taguchi analysis on the collected trial results.
// Steps:
// 1. Compute SNR for each trial.
// 2. Calculate average SNR per orthogonal array row.
// 3. Compute grand mean and total sum of squares.
// 4. Perform ANOVA: factor sum of squares, degrees of freedom, mean squares, F-ratios.
// 5. Determine main effects (average SNR per factor level).
// 6. Compute optimal factor levels based on maximum effect.
// 7. Calculate contribution percentages for each factor.
// Returns an AnalysisResult with all computed statistics, optimal levels, and main effects.
func (e *Experiment) Analyze() AnalysisResult {
	oaRows := len(e.OrthogonalArray)
	grandMean := 0.0
	oaSNR := make([]float64, oaRows)
	for i := 0; i < oaRows; i++ {
		sum := 0.0
		count := 0
		for _, r := range e.Results {
			match := true
			for j, factor := range e.ControlFactors {
				if r.Trial.Control[factor.Name] != factor.Levels[e.OrthogonalArray[i][j]-1] {
					match = false
					break
				}
			}
			if match {
				sum += e.calculateSNR(r.Observations)
				count++
			}
		}
		if count > 0 {
			oaSNR[i] = sum / float64(count)
		} else {
			oaSNR[i] = 0
		}
		grandMean += oaSNR[i]
	}
	grandMean /= float64(oaRows)

	totalSS := 0.0
	for _, sn := range oaSNR {
		totalSS += (sn - grandMean) * (sn - grandMean)
	}

	anova := ANOVAResult{
		FactorSS: make(map[string]float64),
		FactorDF: make(map[string]int),
		FactorMS: make(map[string]float64),
		FactorF:  make(map[string]float64),
	}
	mainEffects := map[string][]float64{}
	snrPerFactor := map[string][]float64{}

	for _, factor := range e.ControlFactors {
		levelMeans := make([]float64, len(factor.Levels))
		levelCounts := make([]int, len(factor.Levels))

		for i := 0; i < oaRows; i++ {
			levelIdx := -1
			for j, f := range e.ControlFactors {
				if f.Name == factor.Name {
					levelIdx = e.OrthogonalArray[i][j] - 1
					break
				}
			}
			if levelIdx >= 0 && levelIdx < len(factor.Levels) {
				levelMeans[levelIdx] += oaSNR[i]
				levelCounts[levelIdx]++
			}
		}

		for li := range levelMeans {
			if levelCounts[li] > 0 {
				levelMeans[li] /= float64(levelCounts[li])
			} else {
				levelMeans[li] = 0
			}
		}

		ss := 0.0
		for li := range factor.Levels {
			ss += float64(levelCounts[li]) * (levelMeans[li] - grandMean) * (levelMeans[li] - grandMean)
		}
		dfs := len(factor.Levels) - 1
		anova.FactorSS[factor.Name] = ss
		anova.FactorDF[factor.Name] = dfs
		mainEffects[factor.Name] = levelMeans
		snrPerFactor[factor.Name] = levelMeans
	}

	errorDF := oaRows - 1
	for _, df := range anova.FactorDF {
		errorDF -= df
	}

	if errorDF < 1 {
		errorDF = 1
	}

	errorSS := totalSS
	for _, ss := range anova.FactorSS {
		errorSS -= ss
	}
	errorMS := errorSS / float64(errorDF)
	anova.ErrorDF = errorDF
	anova.ErrorSS = errorSS
	anova.ErrorMS = errorMS

	for f, msSS := range anova.FactorSS {
		df := anova.FactorDF[f]
		ms := msSS / float64(df)
		anova.FactorMS[f] = ms
		anova.FactorF[f] = ms / errorMS
	}

	optimalLevels := map[string]float64{}
	for _, factor := range e.ControlFactors {
		levels := mainEffects[factor.Name]
		bestLevel := 0
		maxVal := levels[0]
		for i, v := range levels {
			if v > maxVal {
				maxVal = v
				bestLevel = i
			}
		}
		optimalLevels[factor.Name] = factor.Levels[bestLevel]
	}

	totalFactorSS := 0.0
	for _, ss := range anova.FactorSS {
		totalFactorSS += ss
	}
	contributions := map[string]float64{}
	if totalFactorSS > 0 {
		for f, ss := range anova.FactorSS {
			contributions[f] = (ss / totalFactorSS) * 100
		}
	} else {
		for f := range anova.FactorSS {
			contributions[f] = 0
		}
	}

	return AnalysisResult{
		OptimalLevels: optimalLevels,
		SNR:           snrPerFactor,
		MainEffects:   mainEffects,
		Contributions: contributions,
		ANOVA:         anova,
	}
}
