package taguchi

import "testing"

func setupTrialExperiment(arrayName ArrayType, noiseFactors []NoiseFactor) *Experiment[struct{}] {
	factors := []ControlFactor{
		{Name: "A", Levels: []float64{1, 2}},
		{Name: "B", Levels: []float64{10, 20}},
	}
	exp, _ := NewExperimentFromFactors(SmallerTheBetter{}, factors, arrayName, noiseFactors)
	return exp
}

func TestGenerateTrials_Count(t *testing.T) {
	noise := []NoiseFactor{
		{Name: "N1", Levels: []float64{0, 1}},
	}
	exp := setupTrialExperiment(L4, noise)
	trials := exp.GenerateTrials()

	// L4 has 4 rows, 2 noise levels = 4*2 = 8 trials
	want := 8
	if len(trials) != want {
		t.Errorf("GenerateTrials() returned %d trials, want %d", len(trials), want)
	}
}

func TestGenerateTrials_ControlConfigs(t *testing.T) {
	noise := []NoiseFactor{
		{Name: "N1", Levels: []float64{0}},
	}
	exp := setupTrialExperiment(L4, noise)
	trials := exp.GenerateTrials()

	// L4 first 2 columns: {1,1}, {1,2}, {2,1}, {2,2}
	// With A=[1,2], B=[10,20]:
	// Row 0: A=1, B=10
	// Row 1: A=1, B=20
	// Row 2: A=2, B=10
	// Row 3: A=2, B=20
	expectedControls := []map[string]float64{
		{"A": 1, "B": 10},
		{"A": 1, "B": 20},
		{"A": 2, "B": 10},
		{"A": 2, "B": 20},
	}

	if len(trials) != 4 {
		t.Fatalf("GenerateTrials() returned %d trials, want 4", len(trials))
	}

	for i, trial := range trials {
		for key, wantVal := range expectedControls[i] {
			if gotVal := trial.Control[key]; gotVal != wantVal {
				t.Errorf("Trial %d: Control[%q] = %f, want %f", i, key, gotVal, wantVal)
			}
		}
	}
}

func TestGenerateTrials_NoiseCombinations(t *testing.T) {
	noise := []NoiseFactor{
		{Name: "N1", Levels: []float64{0, 1}},
	}
	exp := setupTrialExperiment(L4, noise)
	trials := exp.GenerateTrials()

	// Each OA row should produce 2 trials (one per noise level)
	// Check first 2 trials (from row 0)
	if trials[0].Noise["N1"] != 0 {
		t.Errorf("Trial 0: Noise[N1] = %f, want 0", trials[0].Noise["N1"])
	}
	if trials[1].Noise["N1"] != 1 {
		t.Errorf("Trial 1: Noise[N1] = %f, want 1", trials[1].Noise["N1"])
	}
}

func TestGenerateTrials_NoNoise(t *testing.T) {
	exp := setupTrialExperiment(L4, nil)
	trials := exp.GenerateTrials()

	// No noise factors -> 0 noise combinations -> 0 trials
	// Actually, with no noise factors, generateNoiseCombinations returns 1 trial
	// (the empty combination), so we get 4*1 = 4 trials
	// Let's verify: the helper starts with idx=0, len(noiseFactors)=0, so idx>=len immediately
	// and appends one Trial with empty noise map.
	want := 4
	if len(trials) != want {
		t.Errorf("GenerateTrials() with no noise returned %d trials, want %d", len(trials), want)
	}

	// Noise maps should be empty
	for i, trial := range trials {
		if len(trial.Noise) != 0 {
			t.Errorf("Trial %d: Noise has %d entries, want 0", i, len(trial.Noise))
		}
	}
}

func TestGenerateTrials_MultipleNoise(t *testing.T) {
	noise := []NoiseFactor{
		{Name: "N1", Levels: []float64{0, 1}},
		{Name: "N2", Levels: []float64{100, 200}},
	}
	exp := setupTrialExperiment(L4, noise)
	trials := exp.GenerateTrials()

	// 4 OA rows * (2*2) noise combos = 16 trials
	want := 16
	if len(trials) != want {
		t.Errorf("GenerateTrials() with 2 noise factors returned %d trials, want %d", len(trials), want)
	}
}

func TestGenerateTrials_IDs(t *testing.T) {
	noise := []NoiseFactor{
		{Name: "N1", Levels: []float64{0, 1}},
	}
	exp := setupTrialExperiment(L4, noise)
	trials := exp.GenerateTrials()

	for i, trial := range trials {
		expectedID := i + 1
		if trial.ID != expectedID {
			t.Errorf("Trial %d: ID = %d, want %d", i, trial.ID, expectedID)
		}
	}
}
