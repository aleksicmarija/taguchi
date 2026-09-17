package taguchi

import (
	"errors"
	"math"
	"testing"
)

func near(a, b, tol float64) bool { return math.Abs(a-b) <= tol }

func TestSNRValues(t *testing.T) {
	cases := []struct {
		name string
		snr  SNR
		y    []float64
		want float64
	}{
		{"STB", SmallerTheBetter, []float64{2, 4, 6, 8}, -10 * math.Log10(30)},
		{"STB unit", SmallerTheBetter, []float64{1, 1}, 0},
		{"LTB", LargerTheBetter, []float64{2, 4, 6, 8}, -10 * math.Log10((1.0/4+1.0/16+1.0/36+1.0/64)/4)},
		{"LTB constant", LargerTheBetter, []float64{10, 10}, 20},
		{"NTB target", NominalTheBestTarget(5), []float64{3, 4, 6, 7}, -10 * math.Log10(2.5)},
		{"NTB", NominalTheBest, []float64{4, 6}, 10 * math.Log10(25.0/2-0.5)},
	}
	for _, c := range cases {
		got, err := c.snr.Of(c.y)
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if !near(got, c.want, 1e-9) {
			t.Errorf("%s: got %.6f, want %.6f", c.name, got, c.want)
		}
	}
}

func TestSNRRejects(t *testing.T) {
	nan := math.NaN()
	cases := []struct {
		name string
		snr  SNR
		y    []float64
		want error
	}{
		{"STB empty", SmallerTheBetter, nil, ErrNoObservations},
		{"LTB empty", LargerTheBetter, []float64{}, ErrNoObservations},
		{"NTB empty", NominalTheBest, nil, ErrNoObservations},
		{"NTB target empty", NominalTheBestTarget(1), nil, ErrNoObservations},
		{"STB NaN", SmallerTheBetter, []float64{1, nan}, ErrDomain},
		{"STB Inf", SmallerTheBetter, []float64{math.Inf(1)}, ErrDomain},
		{"LTB zero", LargerTheBetter, []float64{100, 0, 100}, ErrDomain},
		{"LTB negative", LargerTheBetter, []float64{-1, 2}, ErrDomain},
		{"STB all zero", SmallerTheBetter, []float64{0, 0}, ErrUndefinedSNR},
		{"NTB target exact", NominalTheBestTarget(5), []float64{5, 5, 5}, ErrUndefinedSNR},
		{"NTB target NaN", NominalTheBestTarget(nan), []float64{5, 6}, ErrDomain},
		{"NTB one observation", NominalTheBest, []float64{5}, ErrUndefinedSNR},
		{"NTB zero variance", NominalTheBest, []float64{5, 5}, ErrUndefinedSNR},
		{"NTB zero mean", NominalTheBest, []float64{-1, 1}, ErrUndefinedSNR},
		{"NTB mean within one standard error", NominalTheBest, []float64{0.1, -0.1, 0.1, -0.1, 0.2}, ErrUndefinedSNR},
	}
	for _, c := range cases {
		got, err := c.snr.Of(c.y)
		if !errors.Is(err, c.want) {
			t.Errorf("%s: err = %v, want %v", c.name, err, c.want)
		}
		if !math.IsNaN(got) {
			t.Errorf("%s: value %v returned with error", c.name, got)
		}
	}
}

// Scaling every observation by k shifts SmallerTheBetter by -20 log10 k and
// LargerTheBetter by +20 log10 k, leaves NominalTheBest unchanged, and
// shifting observations and target together leaves NominalTheBestTarget
// unchanged. These are the laws that make main-effect differences
// independent of the unit of measurement.
func TestSNRInvariances(t *testing.T) {
	y := []float64{3, 4.5, 6, 7.25, 9}
	const k = 7.5
	scaled := make([]float64, len(y))
	shifted := make([]float64, len(y))
	for i, v := range y {
		scaled[i] = k * v
		shifted[i] = v + 100
	}
	shift := 20 * math.Log10(k)

	check := func(name string, snr SNR, a, b []float64, wantDelta float64) {
		t.Helper()
		x, err1 := snr.Of(a)
		z, err2 := snr.Of(b)
		if err1 != nil || err2 != nil {
			t.Fatalf("%s: %v %v", name, err1, err2)
		}
		if !near(z-x, wantDelta, 1e-9) {
			t.Errorf("%s: delta %.6f, want %.6f", name, z-x, wantDelta)
		}
	}
	check("STB scale", SmallerTheBetter, y, scaled, -shift)
	check("LTB scale", LargerTheBetter, y, scaled, shift)
	check("NTB scale", NominalTheBest, y, scaled, 0)
	x, _ := NominalTheBestTarget(5).Of(y)
	z, _ := NominalTheBestTarget(105).Of(shifted)
	if !near(x, z, 1e-9) {
		t.Errorf("NTB target shift: %.6f vs %.6f", x, z)
	}
}

func TestSNRNames(t *testing.T) {
	for _, c := range []struct {
		snr  SNR
		want string
	}{
		{SmallerTheBetter, "smaller-the-better"},
		{LargerTheBetter, "larger-the-better"},
		{NominalTheBest, "nominal-the-best"},
		{NominalTheBestTarget(2.5), "nominal-the-best, target 2.5"},
	} {
		if c.snr.Name != c.want || c.snr.Of == nil {
			t.Errorf("got %q, want %q", c.snr.Name, c.want)
		}
	}
}
