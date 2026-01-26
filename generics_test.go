package taguchi

import (
	"math"
	"testing"
)

type validFactors struct {
	MaxWorkers []float64
	Algorithm  []float64
}

type validParams struct {
	MaxWorkers float64
	Algorithm  float64
}

type singleFieldFactors struct {
	Speed []float64
}

type mixedFieldsFactors struct {
	Tagged    []float64
	NoSlice   float64
	IntSlice  []int
	unexpored []float64 //nolint:unused
	Another   []float64
}

type tooFewLevels struct {
	X []float64
}

type noSliceFields struct {
	X float64
	Y int
}

func TestFactorsFrom_Valid(t *testing.T) {
	factors, err := factorsFrom(validFactors{
		MaxWorkers: []float64{6, 9, 15, 20},
		Algorithm:  []float64{0, 1},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(factors) != 2 {
		t.Fatalf("expected 2 factors, got %d", len(factors))
	}

	if factors[0].Name != "MaxWorkers" {
		t.Errorf("expected first factor name MaxWorkers, got %s", factors[0].Name)
	}
	expectedLevels := []float64{6, 9, 15, 20}
	if len(factors[0].Levels) != len(expectedLevels) {
		t.Fatalf("expected %d levels, got %d", len(expectedLevels), len(factors[0].Levels))
	}
	for i, v := range expectedLevels {
		if factors[0].Levels[i] != v {
			t.Errorf("level %d: expected %v, got %v", i, v, factors[0].Levels[i])
		}
	}

	if factors[1].Name != "Algorithm" {
		t.Errorf("expected second factor name Algorithm, got %s", factors[1].Name)
	}
	if len(factors[1].Levels) != 2 || factors[1].Levels[0] != 0 || factors[1].Levels[1] != 1 {
		t.Errorf("unexpected Algorithm levels: %v", factors[1].Levels)
	}
}

func TestFactorsFrom_NonStruct(t *testing.T) {
	_, err := factorsFrom(42)
	if err == nil {
		t.Fatal("expected error for non-struct value")
	}
}

func TestFactorsFrom_TooFewLevels(t *testing.T) {
	_, err := factorsFrom(tooFewLevels{X: []float64{5}})
	if err == nil {
		t.Fatal("expected error for single level")
	}
}

func TestFactorsFrom_NoSliceFields(t *testing.T) {
	_, err := factorsFrom(noSliceFields{X: 1.0, Y: 2})
	if err == nil {
		t.Fatal("expected error when no []float64 fields")
	}
}

func TestFactorsFrom_MixedFields(t *testing.T) {
	factors, err := factorsFrom(mixedFieldsFactors{
		Tagged:   []float64{1, 2, 3},
		NoSlice:  99,
		IntSlice: []int{1, 2},
		Another:  []float64{10, 20},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(factors) != 2 {
		t.Fatalf("expected 2 factors (Tagged, Another), got %d", len(factors))
	}
	if factors[0].Name != "Tagged" || factors[1].Name != "Another" {
		t.Errorf("unexpected factor names: %s, %s", factors[0].Name, factors[1].Name)
	}
}

func TestFactorsFrom_FloatLevels(t *testing.T) {
	factors, err := factorsFrom(singleFieldFactors{
		Speed: []float64{1.5, 2.5, 3.5},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(factors) != 1 {
		t.Fatalf("expected 1 factor, got %d", len(factors))
	}
	expected := []float64{1.5, 2.5, 3.5}
	for i, v := range expected {
		if factors[0].Levels[i] != v {
			t.Errorf("level %d: expected %v, got %v", i, v, factors[0].Levels[i])
		}
	}
}

func TestControlAs_Populated(t *testing.T) {
	exp, err := NewExperiment[validFactors, validParams](
		SmallerTheBetter{},
		validFactors{
			MaxWorkers: []float64{6, 9, 15, 20},
			Algorithm:  []float64{0, 1},
		},
		L8,
		nil,
	)
	if err != nil {
		t.Fatalf("NewExperiment failed: %v", err)
	}

	trial := Trial{
		Control: map[string]float64{
			"MaxWorkers": 15,
			"Algorithm":  1,
		},
	}
	params := exp.Params(trial)
	if params.MaxWorkers != 15 {
		t.Errorf("expected MaxWorkers=15, got %v", params.MaxWorkers)
	}
	if params.Algorithm != 1 {
		t.Errorf("expected Algorithm=1, got %v", params.Algorithm)
	}
}

func TestControlAs_MissingKey(t *testing.T) {
	exp, err := NewExperiment[validFactors, validParams](
		SmallerTheBetter{},
		validFactors{
			MaxWorkers: []float64{6, 9, 15, 20},
			Algorithm:  []float64{0, 1},
		},
		L8,
		nil,
	)
	if err != nil {
		t.Fatalf("NewExperiment failed: %v", err)
	}

	trial := Trial{
		Control: map[string]float64{
			"MaxWorkers": 9,
		},
	}
	params := exp.Params(trial)
	if params.MaxWorkers != 9 {
		t.Errorf("expected MaxWorkers=9, got %v", params.MaxWorkers)
	}
	if params.Algorithm != 0 {
		t.Errorf("expected Algorithm=0 (zero value), got %v", params.Algorithm)
	}
}

func TestControlAs_EmptyControl(t *testing.T) {
	exp, err := NewExperiment[validFactors, validParams](
		SmallerTheBetter{},
		validFactors{
			MaxWorkers: []float64{6, 9, 15, 20},
			Algorithm:  []float64{0, 1},
		},
		L8,
		nil,
	)
	if err != nil {
		t.Fatalf("NewExperiment failed: %v", err)
	}

	trial := Trial{Control: map[string]float64{}}
	params := exp.Params(trial)
	if params.MaxWorkers != 0 || params.Algorithm != 0 {
		t.Errorf("expected zero values, got MaxWorkers=%v Algorithm=%v", params.MaxWorkers, params.Algorithm)
	}
}

func TestRoundTrip(t *testing.T) {
	noise := []NoiseFactor{{Name: "Noise", Levels: []float64{0, 1}}}
	exp, err := NewExperiment[validFactors, validParams](
		SmallerTheBetter{},
		validFactors{
			MaxWorkers: []float64{6, 9, 15, 20},
			Algorithm:  []float64{0, 1},
		},
		L8,
		noise,
	)
	if err != nil {
		t.Fatalf("NewExperiment failed: %v", err)
	}

	trials := exp.GenerateTrials()
	if len(trials) == 0 {
		t.Fatal("no trials generated")
	}

	for _, trial := range trials {
		params := exp.Params(trial)

		// Verify the struct values match the map values
		if math.Abs(params.MaxWorkers-trial.Control["MaxWorkers"]) > 1e-10 {
			t.Errorf("trial %d: MaxWorkers mismatch: struct=%v map=%v",
				trial.ID, params.MaxWorkers, trial.Control["MaxWorkers"])
		}
		if math.Abs(params.Algorithm-trial.Control["Algorithm"]) > 1e-10 {
			t.Errorf("trial %d: Algorithm mismatch: struct=%v map=%v",
				trial.ID, params.Algorithm, trial.Control["Algorithm"])
		}

		// Verify values are valid levels
		validWorkers := map[float64]bool{6: true, 9: true, 15: true, 20: true}
		if !validWorkers[params.MaxWorkers] {
			t.Errorf("trial %d: MaxWorkers %v not a valid level", trial.ID, params.MaxWorkers)
		}
		validAlg := map[float64]bool{0: true, 1: true}
		if !validAlg[params.Algorithm] {
			t.Errorf("trial %d: Algorithm %v not a valid level", trial.ID, params.Algorithm)
		}
	}
}

func TestNewExperiment_NonStruct(t *testing.T) {
	_, err := NewExperiment[int, validParams](
		SmallerTheBetter{},
		42,
		L8,
		nil,
	)
	if err == nil {
		t.Fatal("expected error for non-struct factors")
	}
}

func TestNewExperiment_TooFewLevels(t *testing.T) {
	_, err := NewExperiment[tooFewLevels, struct{}](
		SmallerTheBetter{},
		tooFewLevels{X: []float64{5}},
		L8,
		nil,
	)
	if err == nil {
		t.Fatal("expected error for too few levels")
	}
}
