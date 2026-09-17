package taguchi

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"testing"
)

func mustNewArray(t *testing.T, name string, rows [][]int) OrthogonalArray {
	t.Helper()
	a, err := NewOrthogonalArray(name, rows)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func stb(t *testing.T, ys ...float64) []float64 {
	t.Helper()
	out := make([]float64, len(ys))
	for i, y := range ys {
		v, err := SmallerTheBetter.Of([]float64{y})
		if err != nil {
			t.Fatal(err)
		}
		out[i] = v
	}
	return out
}

// Two 2-level factors on the 2x2 full factorial, one observation per run.
// Every expected number is computed here by the textbook formulas, and the
// rounded values are pinned so that the test cannot drift with the code.
func TestAnalyzeSNRHandComputed(t *testing.T) {
	array := mustNewArray(t, "2x2", [][]int{{1, 1}, {1, 2}, {2, 1}, {2, 2}})
	eta := stb(t, 2, 4, 6, 10)

	mean := (eta[0] + eta[1] + eta[2] + eta[3]) / 4
	total := 0.0
	for _, v := range eta {
		total += (v - mean) * (v - mean)
	}
	a1, a2 := (eta[0]+eta[1])/2, (eta[2]+eta[3])/2
	b1, b2 := (eta[0]+eta[2])/2, (eta[1]+eta[3])/2
	ssA := 2*(a1-mean)*(a1-mean) + 2*(a2-mean)*(a2-mean)
	ssB := 2*(b1-mean)*(b1-mean) + 2*(b2-mean)*(b2-mean)
	ssE := total - ssA - ssB

	a, err := AnalyzeSNR(array, eta)
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct {
		name      string
		got, want float64
	}{
		{"Mean", a.Mean, mean},
		{"TotalSS", a.TotalSS, total},
		{"A.SS", a.Effects[0].SS, ssA},
		{"A.SS pinned", a.Effects[0].SS, 76.5732},
		{"B.SS", a.Effects[1].SS, ssB},
		{"B.SS pinned", a.Effects[1].SS, 27.3402},
		{"Error.SS", a.Error.SS, ssE},
		{"Error.SS pinned", a.Error.SS, 0.6270},
		{"A.MS", a.Effects[0].MS, ssA},
		{"Error.MS", a.Error.MS, ssE},
		{"A.F", a.Effects[0].F, ssA / ssE},
		{"A.F pinned", a.Effects[0].F, 122.1328},
		{"B.F", a.Effects[1].F, ssB / ssE},
		{"A.P", a.Effects[0].P, fPValue(ssA/ssE, 1, 1)},
		{"A.Contribution", a.Effects[0].Contribution, 100 * ssA / total},
		{"B.Contribution", a.Effects[1].Contribution, 100 * ssB / total},
		{"Error.Contribution", a.Error.Contribution, 100 * ssE / total},
		{"A.Means[0]", a.Effects[0].Means[0], a1},
		{"A.Means[1]", a.Effects[0].Means[1], a2},
		{"B.Means[1]", a.Effects[1].Means[1], b2},
		{"Predicted", a.PredictedSNR(), mean + (a1 - mean) + (b1 - mean)},
	}
	for _, c := range checks {
		if !near(c.got, c.want, 1e-4) {
			t.Errorf("%s: got %.6f, want %.6f", c.name, c.got, c.want)
		}
	}
	if a.TotalDF != 3 || a.Effects[0].DF != 1 || a.Effects[1].DF != 1 || a.Error.DF != 1 {
		t.Errorf("DF: total %d, A %d, B %d, error %d", a.TotalDF, a.Effects[0].DF, a.Effects[1].DF, a.Error.DF)
	}
	if a.Effects[0].Best != 0 || a.Effects[1].Best != 0 {
		t.Errorf("Best: %d %d, want 0 0", a.Effects[0].Best, a.Effects[1].Best)
	}
	if a.Saturated() {
		t.Error("Saturated with 1 error DF")
	}
	if a.Effects[0].Factor != "column 0" || a.Effects[0].Levels[1] != "1" {
		t.Errorf("default labels: %q %v", a.Effects[0].Factor, a.Effects[0].Levels)
	}
	if e, ok := a.Effect("column 1"); !ok || e.SS != a.Effects[1].SS {
		t.Error("Effect lookup")
	}
	if _, ok := a.Effect("nope"); ok {
		t.Error("Effect lookup found a missing factor")
	}
}

func TestAnalyzeSNRSaturated(t *testing.T) {
	a, err := AnalyzeSNR(L4, []float64{-10, -12, -15, -20})
	if err != nil {
		t.Fatal(err)
	}
	if !a.Saturated() || a.Error.DF != 0 {
		t.Fatalf("Error.DF = %d, want 0", a.Error.DF)
	}
	if a.Error.SS < 0 || a.Error.SS > 1e-9 {
		t.Errorf("Error.SS = %v, want 0", a.Error.SS)
	}
	if !math.IsNaN(a.Error.MS) {
		t.Errorf("Error.MS = %v, want NaN", a.Error.MS)
	}
	sum := 0.0
	for _, e := range a.Effects {
		if !math.IsNaN(e.F) || !math.IsNaN(e.P) {
			t.Errorf("%s: F %v P %v, want NaN", e.Factor, e.F, e.P)
		}
		if e.SS < 0 {
			t.Errorf("%s: negative SS %v", e.Factor, e.SS)
		}
		sum += e.Contribution
	}
	if !near(sum, 100, 1e-9) {
		t.Errorf("contributions sum to %v", sum)
	}
}

func TestPool(t *testing.T) {
	a, err := AnalyzeSNR(L4, []float64{-10, -12, -15, -20})
	if err != nil {
		t.Fatal(err)
	}
	pooled, err := a.Pool("column 2")
	if err != nil {
		t.Fatal(err)
	}
	if a.Saturated() != true || a.Effects[2].Pooled {
		t.Error("Pool modified its receiver")
	}
	if pooled.Saturated() || pooled.Error.DF != 1 || !near(pooled.Error.SS, a.Effects[2].SS, 1e-12) {
		t.Errorf("error term after pooling: SS %v DF %d", pooled.Error.SS, pooled.Error.DF)
	}
	p := pooled.Effects[2]
	if !p.Pooled || !math.IsNaN(p.F) || !math.IsNaN(p.P) || !math.IsNaN(p.Contribution) || math.IsNaN(p.MS) {
		t.Errorf("pooled effect: %+v", p)
	}
	sum := pooled.Error.Contribution
	for _, e := range pooled.Effects[:2] {
		if math.IsNaN(e.F) || e.F <= 0 || e.P <= 0 || e.P >= 1 {
			t.Errorf("%s: F %v P %v", e.Factor, e.F, e.P)
		}
		if !near(e.F, e.MS/pooled.Error.MS, 1e-12) {
			t.Errorf("%s: F %v, want %v", e.Factor, e.F, e.MS/pooled.Error.MS)
		}
		sum += e.Contribution
	}
	if !near(sum, 100, 1e-9) {
		t.Errorf("unpooled contributions plus error sum to %v", sum)
	}
	if !near(pooled.PredictedSNR(), a.Mean+(a.Effects[0].Means[a.Effects[0].Best]-a.Mean)+(a.Effects[1].Means[a.Effects[1].Best]-a.Mean), 1e-12) {
		t.Error("PredictedSNR should exclude pooled factors")
	}
	if _, err := pooled.Pool("column 2"); err == nil {
		t.Error("pooling twice accepted")
	}
	if _, err := a.Pool("nope"); err == nil {
		t.Error("unknown factor accepted")
	}
}

// pseudo returns deterministic values in [-30, -10) for property tests.
func pseudo(n int, seed uint64) []float64 {
	out := make([]float64, n)
	x := seed
	for i := range out {
		x ^= x << 13
		x ^= x >> 7
		x ^= x << 17
		out[i] = -30 + 20*float64(x%1000)/1000
	}
	return out
}

func TestAnalyzeSNRProperties(t *testing.T) {
	cases := []struct {
		array   OrthogonalArray
		errorDF int
	}{
		{L8.Select(0, 1, 2), 4},
		{L8, 0},
		{L9.Select(0, 1), 4},
		{L12.Select(0, 1, 2, 3, 4), 6},
		{L16.Select(0, 1, 2, 3, 4, 5), 9},
		{L18, 2},
	}
	for _, c := range cases {
		snr := pseudo(c.array.Runs(), 0x9E3779B97F4A7C15)
		a, err := AnalyzeSNR(c.array, snr)
		if err != nil {
			t.Fatal(err)
		}
		if a.Error.DF != c.errorDF {
			t.Errorf("%s/%d: Error.DF %d, want %d", c.array.Name(), c.array.Columns(), a.Error.DF, c.errorDF)
		}
		explained := 0.0
		for _, e := range a.Effects {
			explained += e.SS
			// In a balanced design the level means average to the grand mean.
			m := 0.0
			for _, v := range e.Means {
				m += v
			}
			if !near(m/float64(len(e.Means)), a.Mean, 1e-9) {
				t.Errorf("%s/%d %s: level means do not average to the grand mean", c.array.Name(), c.array.Columns(), e.Factor)
			}
		}
		if !near(explained+a.Error.SS, a.TotalSS, 1e-9*a.TotalSS) {
			t.Errorf("%s/%d: SS not additive: %v + %v != %v", c.array.Name(), c.array.Columns(), explained, a.Error.SS, a.TotalSS)
		}
		if a.Error.SS < 0 {
			t.Errorf("%s/%d: negative error SS", c.array.Name(), c.array.Columns())
		}

		// Translation invariance: adding a constant to every SNR changes
		// only the means.
		shifted := make([]float64, len(snr))
		for i, v := range snr {
			shifted[i] = v + 5
		}
		b, err := AnalyzeSNR(c.array, shifted)
		if err != nil {
			t.Fatal(err)
		}
		if !near(b.Mean, a.Mean+5, 1e-9) || !near(b.TotalSS, a.TotalSS, 1e-9) {
			t.Errorf("%s/%d: translation changed the decomposition", c.array.Name(), c.array.Columns())
		}
		for j := range a.Effects {
			if !near(b.Effects[j].SS, a.Effects[j].SS, 1e-9) || b.Effects[j].Best != a.Effects[j].Best {
				t.Errorf("%s/%d %s: translation changed SS or Best", c.array.Name(), c.array.Columns(), a.Effects[j].Factor)
			}
			if !a.Saturated() && !near(b.Effects[j].F, a.Effects[j].F, 1e-9) {
				t.Errorf("%s/%d %s: translation changed F", c.array.Name(), c.array.Columns(), a.Effects[j].Factor)
			}
		}
	}
}

func TestAnalyzeSNRNoVariation(t *testing.T) {
	a, err := AnalyzeSNR(L8.Select(0, 1), []float64{-3, -3, -3, -3, -3, -3, -3, -3})
	if err != nil {
		t.Fatal(err)
	}
	if a.TotalSS != 0 || !math.IsNaN(a.Effects[0].Contribution) || !math.IsNaN(a.Error.Contribution) {
		t.Errorf("no variation: TotalSS %v contributions %v %v", a.TotalSS, a.Effects[0].Contribution, a.Error.Contribution)
	}
}

func TestAnalyzeSNRExactFit(t *testing.T) {
	// A purely additive response on L8 with 3 factors leaves nothing for
	// the interaction columns: the error SS is zero, F is +Inf and p is 0.
	snr := make([]float64, 8)
	for i := range snr {
		snr[i] = -20 + 3*float64(L8.Level(i, 0)) - 2*float64(L8.Level(i, 1)) + 0.5*float64(L8.Level(i, 2))
	}
	a, err := AnalyzeSNR(L8.Select(0, 1, 2), snr)
	if err != nil {
		t.Fatal(err)
	}
	if a.Error.SS > 1e-9 {
		t.Errorf("Error.SS = %v", a.Error.SS)
	}
	for _, e := range a.Effects {
		if !math.IsInf(e.F, 1) || e.P != 0 {
			t.Errorf("%s: F %v P %v", e.Factor, e.F, e.P)
		}
	}
	if a.Effects[0].Best != 1 || a.Effects[1].Best != 0 || a.Effects[2].Best != 1 {
		t.Errorf("Best: %d %d %d", a.Effects[0].Best, a.Effects[1].Best, a.Effects[2].Best)
	}
}

func TestAnalyzeSNRRejects(t *testing.T) {
	if _, err := AnalyzeSNR(L4, []float64{1, 2, 3}); err == nil {
		t.Error("wrong length accepted")
	}
	if _, err := AnalyzeSNR(OrthogonalArray{}, nil); err == nil {
		t.Error("zero array accepted")
	}
	_, err := AnalyzeSNR(L4, []float64{1, math.Inf(1), 3, 4})
	if !errors.Is(err, ErrUndefinedSNR) {
		t.Errorf("infinite SNR: %v", err)
	}
}

// Effects may be reordered, for example ranked by contribution, without
// changing what Best or the report say, and Effect returns a copy.
func TestEffectsAreOrderIndependent(t *testing.T) {
	workers := NewFactor("workers", 1, 4, 16)
	buffer := NewFactor("buffer", 64, 256, 1024)
	design, err := NewDesign(L9, workers, buffer)
	if err != nil {
		t.Fatal(err)
	}
	exp := NewExperiment(design)
	for _, run := range design.Runs() {
		exp.Observe(run, 100/float64(workers.Of(run))+float64(buffer.Of(run))/50)
	}
	a, err := exp.Analyze(SmallerTheBetter)
	if err != nil {
		t.Fatal(err)
	}
	if workers.Best(a) != 16 || buffer.Best(a) != 64 {
		t.Fatalf("before sort: %d %d", workers.Best(a), buffer.Best(a))
	}
	before := a.String()
	slices.Reverse(a.Effects)
	if workers.Best(a) != 16 || buffer.Best(a) != 64 {
		t.Errorf("after sort: %d %d", workers.Best(a), buffer.Best(a))
	}
	if !strings.Contains(a.String(), "BEST SETTINGS: buffer=64 workers=16") {
		t.Errorf("report after sort:\n%s", a)
	}
	// Each run row must still carry the right label under each header.
	for _, report := range []string{before, a.String()} {
		lines := strings.Split(report, "\n")
		header := strings.Fields(lines[3])
		for i, run := range design.Runs() {
			row := strings.Fields(lines[4+i])
			for j, name := range header[1 : len(header)-2] {
				want := map[string]string{"workers": fmt.Sprint(workers.Of(run)), "buffer": fmt.Sprint(buffer.Of(run))}[name]
				if row[1+j] != want {
					t.Errorf("run %d, column %s: %s, want %s", i, name, row[1+j], want)
				}
			}
		}
	}

	e, _ := a.Effect("workers")
	e.Means[0] = 12345
	if e2, _ := a.Effect("workers"); e2.Means[0] == 12345 {
		t.Error("Effect returned a view of the analysis")
	}
}
