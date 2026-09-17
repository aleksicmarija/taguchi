package taguchi

import (
	"math"
	"testing"
)

// Upper critical values of the F distribution from standard tables.
func TestFPValueAgainstTables(t *testing.T) {
	cases := []struct {
		f        float64
		df1, df2 int
		alpha    float64
	}{
		{161.4, 1, 1, 0.05},
		{4.965, 1, 10, 0.05},
		{4.103, 2, 10, 0.05},
		{3.490, 3, 12, 0.05},
		{2.866, 4, 20, 0.05},
		{6.927, 2, 12, 0.01},
		{10.04, 1, 10, 0.01},
	}
	for _, c := range cases {
		p := fPValue(c.f, c.df1, c.df2)
		if !near(p, c.alpha, 5e-4) {
			t.Errorf("P(F(%d,%d) > %v) = %.5f, want %.3f", c.df1, c.df2, c.f, p, c.alpha)
		}
	}
}

func TestFPValueEdges(t *testing.T) {
	if p := fPValue(1, 1, 1); !near(p, 0.5, 1e-12) {
		t.Errorf("median of F(1,1): %v", p)
	}
	if p := fPValue(0, 2, 5); p != 1 {
		t.Errorf("f=0: %v", p)
	}
	if p := fPValue(math.Inf(1), 2, 5); p != 0 {
		t.Errorf("f=inf: %v", p)
	}
	for _, p := range []float64{fPValue(math.NaN(), 1, 1), fPValue(2, 0, 1), fPValue(2, 1, 0)} {
		if !math.IsNaN(p) {
			t.Errorf("want NaN, got %v", p)
		}
	}
}

func TestRegularizedIncompleteBeta(t *testing.T) {
	for _, x := range []float64{0.1, 0.25, 0.5, 0.9} {
		if got := regularizedIncompleteBeta(1, 1, x); !near(got, x, 1e-12) {
			t.Errorf("I_%v(1,1) = %v", x, got)
		}
	}
	if got := regularizedIncompleteBeta(0.5, 0.5, 0.5); !near(got, 0.5, 1e-12) {
		t.Errorf("I_0.5(0.5,0.5) = %v", got)
	}
	a, b, x := 2.5, 7.0, 0.3
	if s := regularizedIncompleteBeta(a, b, x) + regularizedIncompleteBeta(b, a, 1-x); !near(s, 1, 1e-12) {
		t.Errorf("symmetry: %v", s)
	}
}
