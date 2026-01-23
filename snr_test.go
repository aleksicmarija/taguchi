package taguchi

import (
	"math"
	"testing"
)

const tolerance = 1e-4

func almostEqual(a, b, tol float64) bool {
	if math.IsInf(a, 1) && math.IsInf(b, 1) {
		return true
	}
	if math.IsInf(a, -1) && math.IsInf(b, -1) {
		return true
	}
	return math.Abs(a-b) < tol
}

func TestSNRSmallerTheBetter(t *testing.T) {
	obs := []float64{1, 2, 3}
	// msd = (1+4+9)/3 = 14/3
	// SNR = -10*log10(14/3) ≈ -6.6890
	want := -10 * math.Log10(14.0/3.0)
	got := snrSmallerTheBetter(obs)
	if !almostEqual(got, want, tolerance) {
		t.Errorf("snrSmallerTheBetter([1,2,3]) = %f, want %f", got, want)
	}
}

func TestSNRLargerTheBetter(t *testing.T) {
	obs := []float64{2, 4, 6}
	// msd = (1/4 + 1/16 + 1/36)/3 = (0.25 + 0.0625 + 0.027778)/3
	msd := (1.0/4.0 + 1.0/16.0 + 1.0/36.0) / 3.0
	want := -10 * math.Log10(msd)
	got := snrLargerTheBetter(obs)
	if !almostEqual(got, want, tolerance) {
		t.Errorf("snrLargerTheBetter([2,4,6]) = %f, want %f", got, want)
	}
}

func TestSNRNominalTheBest(t *testing.T) {
	obs := []float64{9, 10, 11}
	target := 10.0
	// msd = (1+0+1)/3 = 2/3
	// SNR = -10*log10(2/3) ≈ 1.7609
	want := -10 * math.Log10(2.0 / 3.0)
	got := snrNominalTheBest(obs, target)
	if !almostEqual(got, want, tolerance) {
		t.Errorf("snrNominalTheBest([9,10,11], 10) = %f, want %f", got, want)
	}
}

func TestSNREmptyObservations(t *testing.T) {
	exp := &Experiment{Goal: SmallerTheBetter}
	got := exp.calculateSNR([]float64{})
	if got != 0 {
		t.Errorf("calculateSNR(empty) = %f, want 0", got)
	}
}

func TestSNRSmallerZeroMSD(t *testing.T) {
	// All zeros -> msd=0 -> should return +Inf
	obs := []float64{0, 0, 0}
	got := snrSmallerTheBetter(obs)
	if !math.IsInf(got, 1) {
		t.Errorf("snrSmallerTheBetter([0,0,0]) = %f, want +Inf", got)
	}
}

func TestSNRNominalZeroMSD(t *testing.T) {
	// All observations equal target -> msd=0 -> should return +Inf
	obs := []float64{5, 5, 5}
	got := snrNominalTheBest(obs, 5.0)
	if !math.IsInf(got, 1) {
		t.Errorf("snrNominalTheBest([5,5,5], 5) = %f, want +Inf", got)
	}
}

func TestSNRLargerZeroObservation(t *testing.T) {
	// y=0 should be substituted with 1e-10, not panic
	obs := []float64{0, 1, 2}
	got := snrLargerTheBetter(obs)
	if math.IsNaN(got) || math.IsInf(got, 0) {
		t.Errorf("snrLargerTheBetter([0,1,2]) = %f, want finite value", got)
	}
	// The substitution of 1e-10 makes 1/(1e-10)^2 = 1e20 dominate
	// msd = (1e20 + 1 + 0.25)/3
	// Result should be a very large negative SNR
	if got >= 0 {
		t.Errorf("snrLargerTheBetter([0,1,2]) = %f, expected negative due to near-zero value", got)
	}
}

func TestSNRSingleObservation(t *testing.T) {
	tests := []struct {
		name   string
		goal   OptimizationGoal
		obs    []float64
		target float64
		want   float64
	}{
		{
			name: "SmallerTheBetter single",
			goal: SmallerTheBetter,
			obs:  []float64{5},
			want: -10 * math.Log10(25),
		},
		{
			name: "LargerTheBetter single",
			goal: LargerTheBetter,
			obs:  []float64{5},
			want: -10 * math.Log10(1.0 / 25.0),
		},
		{
			name:   "NominalTheBest single",
			goal:   NominalTheBest,
			obs:    []float64{7},
			target: 5,
			want:   -10 * math.Log10(4),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exp := &Experiment{Goal: tt.goal, Target: tt.target}
			got := exp.calculateSNR(tt.obs)
			if !almostEqual(got, tt.want, tolerance) {
				t.Errorf("calculateSNR(%v) = %f, want %f", tt.obs, got, tt.want)
			}
		})
	}
}
