package taguchi

import "math"

// calculateSNR computes the Signal-to-Noise ratio for a set of observations according to the experiment's goal.
// Returns the SNR value for use in ANOVA and factor effect analysis.
func (e *Experiment) calculateSNR(obs []float64) float64 {
	if len(obs) == 0 {
		return 0
	}

	switch e.Goal {
	case SmallerTheBetter:
		return snrSmallerTheBetter(obs)
	case LargerTheBetter:
		return snrLargerTheBetter(obs)
	case NominalTheBest:
		return snrNominalTheBest(obs, e.Target)
	default:
		return 0
	}
}

// snrSmallerTheBetter calculates the SNR for "smaller-the-better" experiments.
// Formula: -10 * log10(mean(y_i^2))
func snrSmallerTheBetter(obs []float64) float64 {
	msd := 0.0
	for _, y := range obs {
		msd += y * y
	}
	msd /= float64(len(obs))

	if msd == 0 {
		return math.Inf(1)
	}
	return -10 * math.Log10(msd)
}

// snrLargerTheBetter calculates the SNR for "larger-the-better" experiments.
// Formula: -10 * log10(mean(1 / y_i^2))
func snrLargerTheBetter(obs []float64) float64 {
	msd := 0.0
	for _, y := range obs {
		if y == 0 {
			y = 1e-10 // avoid division by zero
		}
		msd += 1 / (y * y)
	}
	msd /= float64(len(obs))
	return -10 * math.Log10(msd)
}

// snrNominalTheBest calculates the SNR for "nominal-the-best" experiments.
// Formula: -10 * log10(mean((y_i - target)^2))
func snrNominalTheBest(obs []float64, target float64) float64 {
	msd := 0.0
	for _, y := range obs {
		msd += (y - target) * (y - target)
	}
	msd /= float64(len(obs))

	if msd == 0 {
		return math.Inf(1)
	}
	return -10 * math.Log10(msd)
}
