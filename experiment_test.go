package taguchi

import (
	"math"
	"testing"
)

const float64EqualityThreshold = 1e-3

func almostEqual(a, b float64) bool {
	return math.Abs(a-b) < float64EqualityThreshold
}

// TestAnalyze_SNR_CombinesObservationsAcrossNoise verifies that SNR is computed
// on the combined observations across all noise conditions for a given OA row,
// not by averaging per-trial SNRs (which is incorrect due to log10 nonlinearity).
//
// Numerical example (SmallerTheBetter: -10*log10(mean(y²))):
//
//	OA row 0 with 2 noise trials: obs [2,4] and [6,8]
//	Buggy:   avg(SNR([2,4]), SNR([6,8])) = avg(-10.0, -16.99) = -13.49
//	Correct: SNR([2,4,6,8]) = -10*log10(mean(4+16+36+64)) = -10*log10(30) ≈ -14.771
func TestAnalyze_SNR_CombinesObservationsAcrossNoise(t *testing.T) {
	factors := []ControlFactor{
		{Name: "A", Levels: []float64{1, 2}},
	}
	oa := [][]int{{1}, {2}} // 2 rows, 1 column
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0, 1}},
	}

	exp, err := NewExperimentFromFactorsUsingArray(SmallerTheBetter{}, factors, oa, noise)
	if err != nil {
		t.Fatalf("NewExperimentFromFactorsUsingArray: %v", err)
	}

	trials := exp.GenerateTrials()
	// trials: (A=1,N=0), (A=1,N=1), (A=2,N=0), (A=2,N=1)
	if len(trials) != 4 {
		t.Fatalf("expected 4 trials, got %d", len(trials))
	}

	// A=1, N=0 → obs [2,4]
	exp.AddResult(trials[0], []float64{2, 4})
	// A=1, N=1 → obs [6,8]
	exp.AddResult(trials[1], []float64{6, 8})
	// A=2, N=0 → obs [1,1]
	exp.AddResult(trials[2], []float64{1, 1})
	// A=2, N=1 → obs [1,1]
	exp.AddResult(trials[3], []float64{1, 1})

	result := exp.Analyze()

	// Expected SNR for A=1: SNR([2,4,6,8]) = -10*log10(mean(4+16+36+64))
	//   = -10*log10((4+16+36+64)/4) = -10*log10(30) ≈ -14.771
	expectedSNR_A1 := -10 * math.Log10(30)

	// Expected SNR for A=2: SNR([1,1,1,1]) = -10*log10(mean(1+1+1+1))
	//   = -10*log10(1) = 0
	expectedSNR_A2 := -10 * math.Log10(1)

	snrA, ok := result.SNR["A"]
	if !ok {
		t.Fatal("SNR for factor A not found")
	}
	if len(snrA) != 2 {
		t.Fatalf("expected 2 SNR levels for A, got %d", len(snrA))
	}

	if !almostEqual(snrA[0], expectedSNR_A1) {
		t.Errorf("SNR[A][0]: got %.4f, want %.4f", snrA[0], expectedSNR_A1)
	}
	if !almostEqual(snrA[1], expectedSNR_A2) {
		t.Errorf("SNR[A][1]: got %.4f, want %.4f", snrA[1], expectedSNR_A2)
	}

	// Optimal level for A should be 2.0 (SNR=0 > SNR≈-14.77)
	if result.OptimalLevels["A"] != 2.0 {
		t.Errorf("OptimalLevels[A]: got %v, want 2.0", result.OptimalLevels["A"])
	}
}

// TestAnalyze_SNR_LargerTheBetter verifies combined-observations SNR for
// LargerTheBetter: -10*log10(mean(1/y²)).
func TestAnalyze_SNR_LargerTheBetter(t *testing.T) {
	factors := []ControlFactor{
		{Name: "A", Levels: []float64{1, 2}},
	}
	oa := [][]int{{1}, {2}}
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0, 1}},
	}

	exp, err := NewExperimentFromFactorsUsingArray(LargerTheBetter{}, factors, oa, noise)
	if err != nil {
		t.Fatalf("NewExperimentFromFactorsUsingArray: %v", err)
	}

	trials := exp.GenerateTrials()
	// A=1, N=0 → obs [2,4]
	exp.AddResult(trials[0], []float64{2, 4})
	// A=1, N=1 → obs [6,8]
	exp.AddResult(trials[1], []float64{6, 8})
	// A=2, N=0 → obs [10,10]
	exp.AddResult(trials[2], []float64{10, 10})
	// A=2, N=1 → obs [10,10]
	exp.AddResult(trials[3], []float64{10, 10})

	result := exp.Analyze()

	// A=1 combined: [2,4,6,8]
	// mean(1/y²) = (1/4 + 1/16 + 1/36 + 1/64)/4
	invSqSum := 1.0/4 + 1.0/16 + 1.0/36 + 1.0/64
	expectedSNR_A1 := -10 * math.Log10(invSqSum/4)

	// A=2 combined: [10,10,10,10]
	// mean(1/y²) = 4*(1/100)/4 = 1/100
	expectedSNR_A2 := -10 * math.Log10(1.0/100)

	snrA := result.SNR["A"]
	if !almostEqual(snrA[0], expectedSNR_A1) {
		t.Errorf("SNR[A][0]: got %.4f, want %.4f", snrA[0], expectedSNR_A1)
	}
	if !almostEqual(snrA[1], expectedSNR_A2) {
		t.Errorf("SNR[A][1]: got %.4f, want %.4f", snrA[1], expectedSNR_A2)
	}

	// A=2 has higher SNR → optimal
	if result.OptimalLevels["A"] != 2.0 {
		t.Errorf("OptimalLevels[A]: got %v, want 2.0", result.OptimalLevels["A"])
	}
}

// TestAnalyze_SNR_NominalTheBest verifies combined-observations SNR for
// NominalTheBest: -10*log10(mean((y-target)²)).
func TestAnalyze_SNR_NominalTheBest(t *testing.T) {
	factors := []ControlFactor{
		{Name: "A", Levels: []float64{1, 2}},
	}
	oa := [][]int{{1}, {2}}
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0, 1}},
	}

	target := 5.0
	exp, err := NewExperimentFromFactorsUsingArray(NominalTheBest{Target: target}, factors, oa, noise)
	if err != nil {
		t.Fatalf("NewExperimentFromFactorsUsingArray: %v", err)
	}

	trials := exp.GenerateTrials()
	// A=1, N=0 → obs [3,4]
	exp.AddResult(trials[0], []float64{3, 4})
	// A=1, N=1 → obs [6,7]
	exp.AddResult(trials[1], []float64{6, 7})
	// A=2, N=0 → obs [5,5]
	exp.AddResult(trials[2], []float64{5, 5})
	// A=2, N=1 → obs [5,5]
	exp.AddResult(trials[3], []float64{5, 5})

	result := exp.Analyze()

	// A=1 combined: [3,4,6,7], deviations from target=5: [-2,-1,1,2]
	// mean((y-5)²) = (4+1+1+4)/4 = 2.5
	expectedSNR_A1 := -10 * math.Log10(2.5)

	// A=2 combined: [5,5,5,5], all exactly on target
	// mean((y-5)²) = 0 → SNR = +Inf
	expectedSNR_A2 := math.Inf(1)

	snrA := result.SNR["A"]
	if !almostEqual(snrA[0], expectedSNR_A1) {
		t.Errorf("SNR[A][0]: got %.4f, want %.4f", snrA[0], expectedSNR_A1)
	}
	if !math.IsInf(snrA[1], 1) {
		t.Errorf("SNR[A][1]: got %.4f, want +Inf", snrA[1])
	}
	_ = expectedSNR_A2

	if result.OptimalLevels["A"] != 2.0 {
		t.Errorf("OptimalLevels[A]: got %v, want 2.0", result.OptimalLevels["A"])
	}
}

// TestAnalyze_SingleTrialPerRow verifies the degenerate case with 1 noise level
// where combined and per-trial SNR would produce the same result.
func TestAnalyze_SingleTrialPerRow(t *testing.T) {
	factors := []ControlFactor{
		{Name: "A", Levels: []float64{1, 2}},
	}
	oa := [][]int{{1}, {2}}
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0}},
	}

	exp, err := NewExperimentFromFactorsUsingArray(SmallerTheBetter{}, factors, oa, noise)
	if err != nil {
		t.Fatalf("NewExperimentFromFactorsUsingArray: %v", err)
	}

	trials := exp.GenerateTrials()
	if len(trials) != 2 {
		t.Fatalf("expected 2 trials, got %d", len(trials))
	}

	exp.AddResult(trials[0], []float64{2, 4})
	exp.AddResult(trials[1], []float64{1, 1})

	result := exp.Analyze()

	// A=1: SNR([2,4]) = -10*log10((4+16)/2) = -10*log10(10) = -10
	expectedSNR_A1 := -10 * math.Log10(10)
	// A=2: SNR([1,1]) = -10*log10(1) = 0
	expectedSNR_A2 := 0.0

	snrA := result.SNR["A"]
	if !almostEqual(snrA[0], expectedSNR_A1) {
		t.Errorf("SNR[A][0]: got %.4f, want %.4f", snrA[0], expectedSNR_A1)
	}
	if !almostEqual(snrA[1], expectedSNR_A2) {
		t.Errorf("SNR[A][1]: got %.4f, want %.4f", snrA[1], expectedSNR_A2)
	}
}

// TestAnalyze_ANOVA_BasicSanity verifies that ANOVA fields are populated and consistent.
func TestAnalyze_ANOVA_BasicSanity(t *testing.T) {
	factors := []ControlFactor{
		{Name: "A", Levels: []float64{1, 2}},
		{Name: "B", Levels: []float64{1, 2}},
	}
	oa := [][]int{{1, 1}, {1, 2}, {2, 1}, {2, 2}}
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0}},
	}

	exp, err := NewExperimentFromFactorsUsingArray(SmallerTheBetter{}, factors, oa, noise)
	if err != nil {
		t.Fatalf("NewExperimentFromFactorsUsingArray: %v", err)
	}

	trials := exp.GenerateTrials()
	// Provide different observation patterns to produce variation
	exp.AddResult(trials[0], []float64{2})  // A=1, B=1
	exp.AddResult(trials[1], []float64{4})  // A=1, B=2
	exp.AddResult(trials[2], []float64{6})  // A=2, B=1
	exp.AddResult(trials[3], []float64{10}) // A=2, B=2

	result := exp.Analyze()

	// Check ANOVA fields exist for both factors
	for _, name := range []string{"A", "B"} {
		if _, ok := result.ANOVA.FactorSS[name]; !ok {
			t.Errorf("ANOVA.FactorSS missing factor %s", name)
		}
		if _, ok := result.ANOVA.FactorDF[name]; !ok {
			t.Errorf("ANOVA.FactorDF missing factor %s", name)
		}
		if _, ok := result.ANOVA.FactorMS[name]; !ok {
			t.Errorf("ANOVA.FactorMS missing factor %s", name)
		}
		if _, ok := result.ANOVA.FactorF[name]; !ok {
			t.Errorf("ANOVA.FactorF missing factor %s", name)
		}
	}

	// DF for 2-level factor should be 1
	if result.ANOVA.FactorDF["A"] != 1 {
		t.Errorf("ANOVA.FactorDF[A]: got %d, want 1", result.ANOVA.FactorDF["A"])
	}
	if result.ANOVA.FactorDF["B"] != 1 {
		t.Errorf("ANOVA.FactorDF[B]: got %d, want 1", result.ANOVA.FactorDF["B"])
	}

	// Contributions should sum to 100%
	totalContrib := 0.0
	for _, c := range result.Contributions {
		totalContrib += c
	}
	if !almostEqual(totalContrib, 100.0) {
		t.Errorf("Contributions sum: got %.4f, want 100.0", totalContrib)
	}

	// SS values should be non-negative
	for name, ss := range result.ANOVA.FactorSS {
		if ss < 0 {
			t.Errorf("ANOVA.FactorSS[%s] is negative: %.4f", name, ss)
		}
	}
}
