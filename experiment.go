package taguchi

import (
	"fmt"
)

// NewExperiment initializes a new generic Taguchi experiment. F is the factors struct type
// (inferred from the factors argument), P is the params struct type for ControlAs.
// arrayName selects a standard orthogonal array (e.g., L4, L8) to generate the trial layout.
func NewExperiment[F any, P any](goal OptimizationGoal, factors F, arrayName ArrayType, noiseFactors []NoiseFactor) (*Experiment[P], error) {
	controlFactors, err := factorsFrom(factors)
	if err != nil {
		return nil, err
	}
	oa, ok := StandardArrays[arrayName]
	if !ok {
		return nil, fmt.Errorf("orthogonal array %s not defined", arrayName)
	}
	if len(controlFactors) > len(oa[0]) {
		return nil, fmt.Errorf("orthogonal array %s cannot accommodate %d factors", arrayName, len(controlFactors))
	}
	return &Experiment[P]{
		ControlFactors:  controlFactors,
		NoiseFactors:    noiseFactors,
		Goal:            goal,
		OrthogonalArray: oa,
		controlAs:       buildControlAs[P](),
	}, nil
}

// NewExperimentUsingArray initializes a new generic Taguchi experiment with a user-provided orthogonal array.
func NewExperimentUsingArray[F any, P any](goal OptimizationGoal, factors F, orthogonalArray [][]int, noiseFactors []NoiseFactor) (*Experiment[P], error) {
	controlFactors, err := factorsFrom(factors)
	if err != nil {
		return nil, err
	}
	if len(orthogonalArray) == 0 {
		return nil, fmt.Errorf("orthogonal array must not be empty")
	}
	if len(controlFactors) > len(orthogonalArray[0]) {
		return nil, fmt.Errorf("orthogonal array cannot accommodate %d factors", len(controlFactors))
	}
	return &Experiment[P]{
		ControlFactors:  controlFactors,
		NoiseFactors:    noiseFactors,
		Goal:            goal,
		OrthogonalArray: orthogonalArray,
		controlAs:       buildControlAs[P](),
	}, nil
}

// NewExperimentFromFactors initializes a Taguchi experiment from a pre-built []Factor slice.
// This is the non-generic constructor for callers who already have []Factor.
func NewExperimentFromFactors(goal OptimizationGoal, controlFactors []ControlFactor, arrayName ArrayType, noiseFactors []NoiseFactor) (*Experiment[struct{}], error) {
	oa, ok := StandardArrays[arrayName]
	if !ok {
		return nil, fmt.Errorf("orthogonal array %s not defined", arrayName)
	}
	if len(controlFactors) > len(oa[0]) {
		return nil, fmt.Errorf("orthogonal array %s cannot accommodate %d factors", arrayName, len(controlFactors))
	}
	return &Experiment[struct{}]{
		ControlFactors:  controlFactors,
		NoiseFactors:    noiseFactors,
		Goal:            goal,
		OrthogonalArray: oa,
	}, nil
}

// NewExperimentFromFactorsUsingArray initializes a Taguchi experiment from a pre-built []Factor slice
// with a user-provided orthogonal array.
func NewExperimentFromFactorsUsingArray(goal OptimizationGoal, controlFactors []ControlFactor, orthogonalArray [][]int, noiseFactors []NoiseFactor) (*Experiment[struct{}], error) {
	if len(orthogonalArray) == 0 {
		return nil, fmt.Errorf("orthogonal array must not be empty")
	}
	if len(controlFactors) > len(orthogonalArray[0]) {
		return nil, fmt.Errorf("orthogonal array cannot accommodate %d factors", len(controlFactors))
	}
	return &Experiment[struct{}]{
		ControlFactors:  controlFactors,
		NoiseFactors:    noiseFactors,
		Goal:            goal,
		OrthogonalArray: orthogonalArray,
	}, nil
}

// Params converts a Trial's Control map into a value of type P using the
// pre-built converter function. P's exported float64 fields are populated from
// the corresponding Control map entries (keyed by field name).
func (e *Experiment[P]) Params(trial Trial) P {
	if e.controlAs == nil {
		var zero P
		return zero
	}
	return e.controlAs(trial)
}

// AddResult records the observations from a completed trial into the experiment's results.
func (e *Experiment[P]) AddResult(trial Trial, observations []float64) {
	e.Results = append(e.Results, TrialResult{
		Trial:        trial,
		Observations: observations,
	})
}

// Analyze performs a full Taguchi analysis on the collected trial results.
func (e *Experiment[P]) Analyze() AnalysisResult {
	oaSNR, grandMean := e.computeOASNR()
	anova, mainEffects, snrPerFactor := e.computeANOVA(oaSNR, grandMean)
	optimalLevels := e.findOptimalLevels(mainEffects)
	contributions := computeContributions(anova)

	return AnalysisResult{
		OptimalLevels: optimalLevels,
		SNR:           snrPerFactor,
		MainEffects:   mainEffects,
		Contributions: contributions,
		ANOVA:         anova,
	}
}

// computeOASNR computes the Signal-to-Noise ratio for each orthogonal array row
// by collecting all observations across noise conditions and computing SNR once
// on the combined set. Returns the per-row SNR values and the grand mean.
func (e *Experiment[P]) computeOASNR() ([]float64, float64) {
	oaRows := len(e.OrthogonalArray)
	oaSNR := make([]float64, oaRows)
	grandMean := 0.0

	for i := 0; i < oaRows; i++ {
		var allObs []float64
		for _, r := range e.Results {
			match := true
			for j, factor := range e.ControlFactors {
				if r.Trial.Control[factor.Name] != factor.Levels[e.OrthogonalArray[i][j]-1] {
					match = false
					break
				}
			}
			if match {
				allObs = append(allObs, r.Observations...)
			}
		}
		if len(allObs) > 0 {
			oaSNR[i] = e.Goal.CalculateSNR(allObs)
		} else {
			oaSNR[i] = 0
		}
		grandMean += oaSNR[i]
	}
	grandMean /= float64(oaRows)

	return oaSNR, grandMean
}

// findOptimalLevels determines the best level for each control factor by
// selecting the level with the highest mean SNR (main effect).
func (e *Experiment[P]) findOptimalLevels(mainEffects map[string][]float64) map[string]float64 {
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
	return optimalLevels
}
