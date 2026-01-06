package taguchi

// GenerateTrials produces all possible trial configurations for the experiment.
func (e *Experiment) GenerateTrials() []Trial {
	// Step 1: Generate all noise combinations
	noiseTrials := e.generateNoiseCombinations()

	// Step 2: Combine noise with orthogonal array control configurations
	finalTrials := e.combineControlAndNoise(noiseTrials)

	return finalTrials
}

// generateNoiseCombinations generates all combinations of noise factors.
// Returns a slice of Trials containing only the Noise field populated (Control is nil).
func (e *Experiment) generateNoiseCombinations() []Trial {
	var trials []Trial
	id := 1

	var helper func(idx int, current map[string]float64)
	helper = func(idx int, current map[string]float64) {
		if idx >= len(e.NoiseFactors) {
			noiseCopy := make(map[string]float64, len(current))
			for k, v := range current {
				noiseCopy[k] = v
			}
			trials = append(trials, Trial{
				ID:      id,
				Control: nil, // to be filled later
				Noise:   noiseCopy,
			})
			id++
			return
		}

		factor := e.NoiseFactors[idx]
		for _, level := range factor.Levels {
			current[factor.Name] = level
			helper(idx+1, current)
		}
	}

	helper(0, map[string]float64{})
	return trials
}

// combineControlAndNoise takes a list of noise-only trials and combines them with all control factor configurations
// defined by the orthogonal array. Returns a slice of fully defined Trials.
func (e *Experiment) combineControlAndNoise(noiseTrials []Trial) []Trial {
	var finalTrials []Trial
	id := 1 // reset ID for full trial list

	for _, row := range e.OrthogonalArray {
		controlConfig := e.getControlConfig(row)

		for _, noiseTrial := range noiseTrials {
			t := Trial{
				ID:      id,
				Control: controlConfig,
				Noise:   noiseTrial.Noise,
			}
			finalTrials = append(finalTrials, t)
			id++
		}
	}

	return finalTrials
}

// getControlConfig converts a single orthogonal array row into a map of control factor names to levels.
func (e *Experiment) getControlConfig(row []int) map[string]float64 {
	controlConfig := make(map[string]float64, len(e.ControlFactors))
	for j, factor := range e.ControlFactors {
		levelIndex := row[j] - 1 // orthogonal array indices are 1-based
		controlConfig[factor.Name] = factor.Levels[levelIndex]
	}
	return controlConfig
}
