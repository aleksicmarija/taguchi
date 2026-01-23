package taguchi

import (
	"math"
	"testing"
)

func TestNewExperiment_Valid(t *testing.T) {
	factors := []Factor{
		{Name: "A", Levels: []float64{1, 2}},
		{Name: "B", Levels: []float64{10, 20}},
	}
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0}},
	}

	exp, err := NewExperiment(SmallerTheBetter, 0, 0.1, factors, "L4", noise)
	if err != nil {
		t.Fatalf("NewExperiment() returned error: %v", err)
	}
	if exp == nil {
		t.Fatal("NewExperiment() returned nil experiment")
	}
	if exp.Goal != SmallerTheBetter {
		t.Errorf("Goal = %v, want SmallerTheBetter", exp.Goal)
	}
	if len(exp.ControlFactors) != 2 {
		t.Errorf("ControlFactors count = %d, want 2", len(exp.ControlFactors))
	}
	if len(exp.OrthogonalArray) != 4 {
		t.Errorf("OrthogonalArray rows = %d, want 4", len(exp.OrthogonalArray))
	}
}

func TestNewExperiment_InvalidArray(t *testing.T) {
	factors := []Factor{
		{Name: "A", Levels: []float64{1, 2}},
	}
	_, err := NewExperiment(SmallerTheBetter, 0, 0, factors, "L99", nil)
	if err == nil {
		t.Error("NewExperiment() with invalid array should return error")
	}
}

func TestNewExperiment_TooManyFactors(t *testing.T) {
	// L4 has 3 columns, so 4 factors should fail
	factors := []Factor{
		{Name: "A", Levels: []float64{1, 2}},
		{Name: "B", Levels: []float64{1, 2}},
		{Name: "C", Levels: []float64{1, 2}},
		{Name: "D", Levels: []float64{1, 2}},
	}
	_, err := NewExperiment(SmallerTheBetter, 0, 0, factors, "L4", nil)
	if err == nil {
		t.Error("NewExperiment() with too many factors should return error")
	}
}

func TestAddResult(t *testing.T) {
	factors := []Factor{
		{Name: "A", Levels: []float64{1, 2}},
	}
	exp, _ := NewExperiment(SmallerTheBetter, 0, 0, factors, "L4", nil)

	trial := Trial{ID: 1, Control: map[string]float64{"A": 1}}
	exp.AddResult(trial, []float64{1.0, 2.0, 3.0})

	if len(exp.Results) != 1 {
		t.Fatalf("Results count = %d, want 1", len(exp.Results))
	}
	if len(exp.Results[0].Observations) != 3 {
		t.Errorf("Observations count = %d, want 3", len(exp.Results[0].Observations))
	}

	exp.AddResult(trial, []float64{4.0})
	if len(exp.Results) != 2 {
		t.Errorf("Results count after second add = %d, want 2", len(exp.Results))
	}
}

func TestAnalyze_L4_SmallerTheBetter(t *testing.T) {
	factors := []Factor{
		{Name: "A", Levels: []float64{1, 2}},
		{Name: "B", Levels: []float64{10, 20}},
	}
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0}},
	}

	exp, err := NewExperiment(SmallerTheBetter, 0, 0, factors, "L4", noise)
	if err != nil {
		t.Fatalf("NewExperiment() error: %v", err)
	}

	trials := exp.GenerateTrials()

	// Observations: use A+B as the response
	// Row 0: A=1, B=10 -> obs=11
	// Row 1: A=1, B=20 -> obs=21
	// Row 2: A=2, B=10 -> obs=12
	// Row 3: A=2, B=20 -> obs=22
	for _, trial := range trials {
		obs := trial.Control["A"] + trial.Control["B"]
		exp.AddResult(trial, []float64{obs})
	}

	result := exp.Analyze()

	// Optimal: smallest values -> A=1, B=10
	if result.OptimalLevels["A"] != 1 {
		t.Errorf("OptimalLevels[A] = %f, want 1", result.OptimalLevels["A"])
	}
	if result.OptimalLevels["B"] != 10 {
		t.Errorf("OptimalLevels[B] = %f, want 10", result.OptimalLevels["B"])
	}

	// Contributions should sum to ~100%
	totalContrib := 0.0
	for _, c := range result.Contributions {
		totalContrib += c
	}
	if !almostEqual(totalContrib, 100.0, 0.1) {
		t.Errorf("Contributions sum = %f, want ~100", totalContrib)
	}

	// ANOVA: degrees of freedom
	// Each factor has 2 levels -> DF=1
	if result.ANOVA.FactorDF["A"] != 1 {
		t.Errorf("ANOVA.FactorDF[A] = %d, want 1", result.ANOVA.FactorDF["A"])
	}
	if result.ANOVA.FactorDF["B"] != 1 {
		t.Errorf("ANOVA.FactorDF[B] = %d, want 1", result.ANOVA.FactorDF["B"])
	}

	// ANOVA: SS should be non-negative
	if result.ANOVA.FactorSS["A"] < 0 {
		t.Errorf("ANOVA.FactorSS[A] = %f, want >= 0", result.ANOVA.FactorSS["A"])
	}
	if result.ANOVA.FactorSS["B"] < 0 {
		t.Errorf("ANOVA.FactorSS[B] = %f, want >= 0", result.ANOVA.FactorSS["B"])
	}

	// B has larger effect than A (range of 10 vs 1), so B should contribute more
	if result.Contributions["B"] < result.Contributions["A"] {
		t.Errorf("B contribution (%f) should be > A contribution (%f)",
			result.Contributions["B"], result.Contributions["A"])
	}

	// Verify main effects exist for both factors
	if len(result.MainEffects["A"]) != 2 {
		t.Errorf("MainEffects[A] has %d levels, want 2", len(result.MainEffects["A"]))
	}
	if len(result.MainEffects["B"]) != 2 {
		t.Errorf("MainEffects[B] has %d levels, want 2", len(result.MainEffects["B"]))
	}
}

func TestAnalyze_L9_NominalTheBest(t *testing.T) {
	factors := []Factor{
		{Name: "X", Levels: []float64{1, 2, 3}},
		{Name: "Y", Levels: []float64{10, 20, 30}},
		{Name: "Z", Levels: []float64{100, 200, 300}},
	}
	noise := []NoiseFactor{
		{Name: "N", Levels: []float64{0}},
	}

	target := 50.0
	exp, err := NewExperiment(NominalTheBest, target, 0, factors, "L9", noise)
	if err != nil {
		t.Fatalf("NewExperiment() error: %v", err)
	}

	trials := exp.GenerateTrials()

	// Response: X + Y (Z is irrelevant noise that adds variability)
	// Target = 50, so best is X=3 + Y=30 = 33? No...
	// Let's make response = X*10 + Y, target=50
	// X=1,Y=10 -> 20; X=2,Y=20 -> 40; X=3,Y=30 -> 60
	// Closest to 50: X=2,Y=30->50 (exact match gets +Inf SNR)
	// But with L9 layout we don't control exact combos.
	// Simpler: response = X*Y, and ignore Z
	// Just verify the analysis runs and produces valid output
	for _, trial := range trials {
		obs := trial.Control["X"] * trial.Control["Y"]
		exp.AddResult(trial, []float64{obs})
	}

	result := exp.Analyze()

	// Verify optimal levels exist for all factors
	for _, f := range []string{"X", "Y", "Z"} {
		if _, ok := result.OptimalLevels[f]; !ok {
			t.Errorf("OptimalLevels missing factor %q", f)
		}
	}

	// Contributions should sum to ~100%
	totalContrib := 0.0
	for _, c := range result.Contributions {
		totalContrib += c
	}
	if !almostEqual(totalContrib, 100.0, 0.1) {
		t.Errorf("Contributions sum = %f, want ~100", totalContrib)
	}

	// Each factor should have 3 levels in main effects
	for _, f := range []string{"X", "Y", "Z"} {
		if len(result.MainEffects[f]) != 3 {
			t.Errorf("MainEffects[%s] has %d levels, want 3", f, len(result.MainEffects[f]))
		}
	}

	// ANOVA DF for 3-level factor = 2
	for _, f := range []string{"X", "Y", "Z"} {
		if result.ANOVA.FactorDF[f] != 2 {
			t.Errorf("ANOVA.FactorDF[%s] = %d, want 2", f, result.ANOVA.FactorDF[f])
		}
	}
}

func TestAnalyze_MainEffects(t *testing.T) {
	factors := []Factor{
		{Name: "A", Levels: []float64{1, 2}},
		{Name: "B", Levels: []float64{10, 20}},
	}

	exp, _ := NewExperiment(SmallerTheBetter, 0, 0, factors, "L4", nil)
	trials := exp.GenerateTrials()

	for _, trial := range trials {
		obs := trial.Control["A"] + trial.Control["B"]
		exp.AddResult(trial, []float64{obs})
	}

	result := exp.Analyze()

	// Main effects for A:
	// Level 1 (A=1): rows with A=1 -> obs 11, 21 -> SNR mean of those rows
	// Level 2 (A=2): rows with A=2 -> obs 12, 22 -> SNR mean of those rows
	// A level 1 should have higher SNR (smaller values -> higher SNR in SmallerTheBetter)
	if result.MainEffects["A"][0] <= result.MainEffects["A"][1] {
		t.Errorf("MainEffects[A]: level 1 (%f) should have higher SNR than level 2 (%f)",
			result.MainEffects["A"][0], result.MainEffects["A"][1])
	}

	// Main effects for B: level 1 (B=10) should have higher SNR than level 2 (B=20)
	if result.MainEffects["B"][0] <= result.MainEffects["B"][1] {
		t.Errorf("MainEffects[B]: level 1 (%f) should have higher SNR than level 2 (%f)",
			result.MainEffects["B"][0], result.MainEffects["B"][1])
	}
}

func TestAnalyze_Contributions(t *testing.T) {
	factors := []Factor{
		{Name: "A", Levels: []float64{1, 2}},
		{Name: "B", Levels: []float64{10, 20}},
		{Name: "C", Levels: []float64{100, 200}},
	}

	exp, _ := NewExperiment(SmallerTheBetter, 0, 0, factors, "L4", nil)
	trials := exp.GenerateTrials()

	// Response dominated by C (largest range)
	for _, trial := range trials {
		obs := trial.Control["A"] + trial.Control["B"] + trial.Control["C"]
		exp.AddResult(trial, []float64{obs})
	}

	result := exp.Analyze()

	totalContrib := 0.0
	for _, c := range result.Contributions {
		totalContrib += c
	}
	if !almostEqual(totalContrib, 100.0, 0.1) {
		t.Errorf("Contributions sum = %f, want ~100", totalContrib)
	}

	// All contributions should be non-negative
	for f, c := range result.Contributions {
		if c < 0 {
			t.Errorf("Contributions[%s] = %f, want >= 0", f, c)
		}
	}

	// C should have the largest contribution (dominates response)
	if result.Contributions["C"] < result.Contributions["A"] ||
		result.Contributions["C"] < result.Contributions["B"] {
		t.Errorf("C contribution (%f) should dominate, got A=%f, B=%f, C=%f",
			result.Contributions["C"],
			result.Contributions["A"],
			result.Contributions["B"],
			result.Contributions["C"])
	}

	// Verify no NaN values in main effects
	for f, effects := range result.MainEffects {
		for i, v := range effects {
			if math.IsNaN(v) {
				t.Errorf("MainEffects[%s][%d] is NaN", f, i)
			}
		}
	}
}
