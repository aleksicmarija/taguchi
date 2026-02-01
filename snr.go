package taguchi

import "math"

// CalculateSNR computes the Signal-to-Noise ratio for "smaller-the-better" experiments.
// Formula: -10 * log10(mean(y_i^2))
func (s SmallerTheBetter) CalculateSNR(obs []float64) float64 {
	if len(obs) == 0 {
		return 0
	}
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

// String returns the human-readable name for the SmallerTheBetter goal.
func (s SmallerTheBetter) String() string {
	return "Smaller-the-Better"
}

// CalculateSNR computes the Signal-to-Noise ratio for "larger-the-better" experiments.
// Formula: -10 * log10(mean(1/y_i^2))
func (l LargerTheBetter) CalculateSNR(obs []float64) float64 {
	if len(obs) == 0 {
		return 0
	}
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

// String returns the human-readable name for the LargerTheBetter goal.
func (l LargerTheBetter) String() string {
	return "Larger-the-Better"
}

// CalculateSNR computes the Signal-to-Noise ratio for "nominal-the-best" experiments.
// Formula: -10 * log10(mean((y_i - Target)^2))
func (n NominalTheBest) CalculateSNR(obs []float64) float64 {
	if len(obs) == 0 {
		return 0
	}
	msd := 0.0
	for _, y := range obs {
		msd += (y - n.Target) * (y - n.Target)
	}
	msd /= float64(len(obs))

	if msd == 0 {
		return math.Inf(1)
	}
	return -10 * math.Log10(msd)
}

// String returns the human-readable name for the NominalTheBest goal.
func (n NominalTheBest) String() string {
	return "Nominal-the-Best"
}
