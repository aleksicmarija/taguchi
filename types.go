package taguchi

// OptimizationGoal defines the type of quality characteristic being optimized.
// It is used to determine how the Signal-to-Noise (SNR) ratio is calculated for trials.
type OptimizationGoal int

const (
	// SmallerTheBetter means the goal is to minimize the response variable.
	SmallerTheBetter OptimizationGoal = iota

	// LargerTheBetter means the goal is to maximize the response variable.
	LargerTheBetter

	// NominalTheBest means the goal is to achieve a target value with minimal deviation.
	NominalTheBest
)

// String provides a human-readable representation of the OptimizationGoal.
// Useful for logging and displaying experiment settings.
func (o OptimizationGoal) String() string {
	switch o {
	case SmallerTheBetter:
		return "Smaller-the-Better"
	case LargerTheBetter:
		return "Larger-the-Better"
	case NominalTheBest:
		return "Nominal-the-Best"
	default:
		return "Unknown"
	}
}

// Factor represents a controllable input variable in the experiment.
// Name: Identifier for the factor (e.g., "NumThreads").
// Levels: A slice of possible numeric values that this factor can take.
type Factor struct {
	Name   string
	Levels []float64
}

// NoiseFactor represents an uncontrollable input variable (noise) in the experiment.
// Name: Identifier for the noise factor (e.g., "CPU Load").
// Levels: A slice of numeric levels representing different environmental conditions.
type NoiseFactor struct {
	Name   string
	Levels []float64
}

// Trial represents a single experimental run combining a specific control and noise configuration.
// ID: Unique identifier for the trial.
// Control: Mapping from factor names to their selected levels for this trial.
// Noise: Mapping from noise factor names to their levels during the trial.
type Trial struct {
	ID      int
	Control map[string]float64
	Noise   map[string]float64
}

// TrialResult stores the observed outcomes from a trial.
// Trial: The trial configuration that produced these observations.
// Observations: Measured results for this trial (e.g., latency measurements).
type TrialResult struct {
	Trial        Trial
	Observations []float64
}

// AnalysisResult stores the results of analyzing all experimental trials.
// OptimalLevels: Maps each control factor to its best-performing level.
// SNR: Signal-to-noise ratios for each factor's levels.
// MainEffects: Average SNR per factor level, showing the effect of each factor.
// Contributions: Percentage contribution of each factor to overall variability.
// ANOVA: Detailed ANOVA statistics including SS, DF, MS, and F-ratio for factors.
type AnalysisResult struct {
	OptimalLevels map[string]float64
	SNR           map[string][]float64
	MainEffects   map[string][]float64
	Contributions map[string]float64
	ANOVA         ANOVAResult
}

// ANOVAResult stores detailed ANOVA calculations for the experiment.
// FactorSS: Sum of squares for each factor.
// FactorDF: Degrees of freedom for each factor.
// FactorMS: Mean square values for each factor.
// FactorF: F-ratio for each factor.
// ErrorSS: Sum of squares for residual/error.
// ErrorDF: Degrees of freedom for residual/error.
// ErrorMS: Mean square error.
// PooledFactors: List of factors that were pooled together during analysis (optional).
type ANOVAResult struct {
	FactorSS      map[string]float64
	FactorDF      map[string]int
	FactorMS      map[string]float64
	FactorF       map[string]float64
	ErrorSS       float64
	ErrorDF       int
	ErrorMS       float64
	PooledFactors []string
}

// Experiment encapsulates all the configuration and results for a Taguchi experiment.
// ControlFactors: Factors we can manipulate.
// NoiseFactors: Uncontrollable environmental factors.
// Goal: Optimization goal (Smaller, Larger, or Nominal).
// Target: Target value (used for NominalTheBest).
// OrthogonalArray: Predefined L4/L8/L9/etc. orthogonal array for trial combinations.
// Results: Collection of TrialResults after experiments.
// PoolingThreshold: Threshold used for pooling small-effect factors in ANOVA.
type Experiment struct {
	ControlFactors   []Factor
	NoiseFactors     []NoiseFactor
	Goal             OptimizationGoal
	Target           float64
	OrthogonalArray  [][]int
	Results          []TrialResult
	PoolingThreshold float64
}
