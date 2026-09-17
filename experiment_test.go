package taguchi

import (
	"errors"
	"math"
	"strings"
	"sync"
	"testing"
)

// twoRuns is the smallest design: one 2-level factor on a 2-run array.
func twoRuns(t *testing.T) (*Design, *Factor[int]) {
	t.Helper()
	array := mustNewArray(t, "2x1", [][]int{{1}, {2}})
	a := NewFactor("A", 1, 2)
	design, err := NewDesign(array, a)
	if err != nil {
		t.Fatal(err)
	}
	return design, a
}

// The SNR of a run is computed on all its observations pooled together,
// not by averaging the SNR of each noise condition; the log makes those
// differ. Observations [2,4] and [6,8] pool to -10 log10(30) ≈ -14.77,
// whereas averaging per-condition SNRs would give -13.49.
func TestAnalyzePoolsObservationsAcrossConditions(t *testing.T) {
	design, a := twoRuns(t)
	exp := NewExperiment(design)
	runs := design.Runs()
	exp.Observe(runs[0], 2, 4)
	exp.Observe(runs[0], 6, 8)
	exp.Observe(runs[1], 1, 1)
	exp.Observe(runs[1], 1, 1)

	an, err := exp.Analyze(SmallerTheBetter)
	if err != nil {
		t.Fatal(err)
	}
	if !near(an.Runs[0].SNR, -10*math.Log10(30), 1e-9) || an.Runs[0].N != 4 {
		t.Errorf("run 0: %+v", an.Runs[0])
	}
	if !near(an.Runs[1].SNR, 0, 1e-9) || an.Runs[1].N != 4 {
		t.Errorf("run 1: %+v", an.Runs[1])
	}
	e := an.Effects[0]
	if e.Factor != "A" || e.Levels[0] != "1" || e.Levels[1] != "2" {
		t.Errorf("labels: %q %v", e.Factor, e.Levels)
	}
	if !near(e.Means[0], -10*math.Log10(30), 1e-9) || !near(e.Means[1], 0, 1e-9) || e.Best != 1 {
		t.Errorf("means %v best %d", e.Means, e.Best)
	}
	if a.Best(an) != 2 {
		t.Errorf("Best = %d", a.Best(an))
	}
	if !an.Saturated() {
		t.Error("one factor on two runs leaves no error DF")
	}
}

func TestAnalyzeLargerTheBetterAndTarget(t *testing.T) {
	design, a := twoRuns(t)
	exp := NewExperiment(design)
	runs := design.Runs()
	exp.Observe(runs[0], 2, 4, 6, 8)
	exp.Observe(runs[1], 10, 10, 10, 10)
	an, err := exp.Analyze(LargerTheBetter)
	if err != nil {
		t.Fatal(err)
	}
	want0 := -10 * math.Log10((1.0/4+1.0/16+1.0/36+1.0/64)/4)
	if !near(an.Effects[0].Means[0], want0, 1e-9) || !near(an.Effects[0].Means[1], 20, 1e-9) || a.Best(an) != 2 {
		t.Errorf("LTB: %v best %d", an.Effects[0].Means, a.Best(an))
	}

	exp = NewExperiment(design)
	exp.Observe(runs[0], 3, 4, 6, 7)
	exp.Observe(runs[1], 4, 6, 4, 6)
	an, err = exp.Analyze(NominalTheBestTarget(5))
	if err != nil {
		t.Fatal(err)
	}
	if !near(an.Effects[0].Means[0], -10*math.Log10(2.5), 1e-9) || !near(an.Effects[0].Means[1], 0, 1e-9) || a.Best(an) != 2 {
		t.Errorf("NTB target: %v best %d", an.Effects[0].Means, a.Best(an))
	}
}

func TestAnalyzeRefusesMissingRuns(t *testing.T) {
	design, err := NewDesign(L4, NewFactor("A", 1, 2), NewFactor("B", 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	exp := NewExperiment(design)
	runs := design.Runs()
	exp.Observe(runs[0], 100)
	exp.Observe(runs[2], 100)
	_, err = exp.Analyze(SmallerTheBetter)
	if !errors.Is(err, ErrNoObservations) || !strings.Contains(err.Error(), "[1 3]") {
		t.Errorf("missing runs: %v", err)
	}
}

func TestAnalyzeNamesTheRunAnSNRRejects(t *testing.T) {
	design, err := NewDesign(L4, NewFactor("A", 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	exp := NewExperiment(design)
	for i, run := range design.Runs() {
		if i == 2 {
			exp.Observe(run, 5, 0)
		} else {
			exp.Observe(run, 5, 6)
		}
	}
	_, err = exp.Analyze(LargerTheBetter)
	if !errors.Is(err, ErrDomain) || !strings.HasPrefix(err.Error(), "run 2:") {
		t.Errorf("domain error: %v", err)
	}
	if _, err := exp.Analyze(SNR{}); err == nil {
		t.Error("zero SNR accepted")
	}
}

func TestExperimentObservations(t *testing.T) {
	design, _ := twoRuns(t)
	exp := NewExperiment(design)
	run := design.Run(1)
	exp.Observe(run, 1, 2)
	exp.Observe(run)
	exp.Observe(run, 3)
	got := exp.Observations(run)
	if len(got) != 3 || got[2] != 3 {
		t.Errorf("observations: %v", got)
	}
	got[0] = 99
	if exp.Observations(run)[0] != 1 {
		t.Error("Observations returned the internal slice")
	}
	if exp.Design() != design {
		t.Error("Design")
	}
	other, _ := twoRuns(t)
	mustPanic(t, "run from another design", func() { exp.Observe(other.Run(0), 1) })
	mustPanic(t, "zero run", func() { exp.Observations(Run{}) })
	mustPanic(t, "nil design", func() { NewExperiment(nil) })
}

func TestExperimentConcurrentObserve(t *testing.T) {
	design, err := NewDesign(L8, NewFactor("A", 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	exp := NewExperiment(design)
	var wg sync.WaitGroup
	for g := 0; g < 50; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for _, run := range design.Runs() {
				exp.Observe(run, float64(g+1), float64(g+2))
			}
		}(g)
	}
	wg.Wait()
	for _, run := range design.Runs() {
		if n := len(exp.Observations(run)); n != 100 {
			t.Errorf("%v: %d observations, want 100", run, n)
		}
	}
	if _, err := exp.Analyze(SmallerTheBetter); err != nil {
		t.Fatal(err)
	}
}

// End to end with typed factors and a noise loop: an additive model with a
// known optimum is recovered, with names and level labels in the result.
func TestExperimentEndToEnd(t *testing.T) {
	workers := NewFactor("workers", 1, 4, 16)
	algo := NewFactor("algorithm", quick, radix, quick)
	buffer := NewFactor("buffer", 64, 256, 1024)
	design, err := NewDesign(L9, workers, algo, buffer)
	if err != nil {
		t.Fatal(err)
	}
	latency := func(w int, a algorithm, b int, noise float64) float64 {
		cost := 100.0 / float64(w)
		if a == radix {
			cost *= 0.5
		}
		cost += float64(b) / 128
		return cost * noise
	}
	exp := NewExperiment(design)
	for _, run := range design.Runs() {
		for _, noise := range []float64{0.9, 1.0, 1.3} {
			exp.Observe(run, latency(workers.Of(run), algo.Of(run), buffer.Of(run), noise))
		}
	}
	an, err := exp.Analyze(SmallerTheBetter)
	if err != nil {
		t.Fatal(err)
	}
	if workers.Best(an) != 16 || algo.Best(an) != radix || buffer.Best(an) != 64 {
		t.Errorf("best: %d %v %d", workers.Best(an), algo.Best(an), buffer.Best(an))
	}
	e, ok := an.Effect("algorithm")
	if !ok || e.Levels[1] != "radix" || e.Best != 1 {
		t.Errorf("algorithm effect: %+v", e)
	}
	if an.Error.DF != 2 || an.Saturated() {
		t.Errorf("Error.DF = %d", an.Error.DF)
	}
	if an.SNR.Name != "smaller-the-better" {
		t.Errorf("SNR recorded as %q", an.SNR.Name)
	}
	for _, r := range an.Runs {
		if r.N != 3 {
			t.Errorf("run %d: N = %d", r.Row, r.N)
		}
	}
	if w, _ := an.Effect("workers"); w.Contribution < 50 {
		t.Errorf("workers should dominate: %+v", w)
	}
}

func TestBalanced(t *testing.T) {
	design, _ := twoRuns(t)
	exp := NewExperiment(design)
	exp.Observe(design.Run(0), 1, 2, 3)
	exp.Observe(design.Run(1), 4, 5, 6)
	an, err := exp.Analyze(SmallerTheBetter)
	if err != nil {
		t.Fatal(err)
	}
	if !an.Balanced() || strings.Contains(an.String(), "equally precise") {
		t.Error("balanced experiment reported as unbalanced")
	}
	exp.Observe(design.Run(1), 7, 8)
	an, err = exp.Analyze(SmallerTheBetter)
	if err != nil {
		t.Fatal(err)
	}
	if an.Balanced() || !strings.Contains(an.String(), "between 3 and 5 observations") {
		t.Errorf("unbalanced experiment: Balanced=%v report:\n%s", an.Balanced(), an)
	}
	pure, err := AnalyzeSNR(L4.Select(0), []float64{1, 2, 3, 4})
	if err != nil {
		t.Fatal(err)
	}
	if !pure.Balanced() {
		t.Error("AnalyzeSNR result should count as balanced")
	}
}
