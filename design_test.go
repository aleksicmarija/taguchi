package taguchi

import (
	"strings"
	"testing"
)

type algorithm int

const (
	quick algorithm = iota
	radix
)

func (a algorithm) String() string { return [...]string{"quick", "radix"}[a] }

func TestNewDesignRejectsLevelMismatch(t *testing.T) {
	// A 3-level factor on a 2-level column used to run two levels and
	// report the third, never executed, as the best.
	_, err := NewDesign(L8, NewFactor("A", 1, 2), NewFactor("B", 10, 20, 30))
	if err == nil || !strings.Contains(err.Error(), "levels") {
		t.Errorf("3-level factor on L8: %v", err)
	}
	// A 2-level factor on a 3-level column used to panic during trial
	// generation.
	_, err = NewDesign(L18, NewFactor("A", 1, 2), NewFactor("B", 1, 2))
	if err == nil || !strings.Contains(err.Error(), "column 1") {
		t.Errorf("2-level factor on L18 column 1: %v", err)
	}
	if _, err := NewDesign(L18, NewFactor("A", 1, 2), NewFactor("B", 1, 2, 3)); err != nil {
		t.Errorf("matching levels rejected: %v", err)
	}
}

func TestNewDesignRejects(t *testing.T) {
	two := func(name string) *Factor[int] { return NewFactor(name, 1, 2) }
	cases := []struct {
		name    string
		array   OrthogonalArray
		factors []AnyFactor
	}{
		{"zero array", OrthogonalArray{}, []AnyFactor{two("A")}},
		{"no factors", L4, nil},
		{"too many factors", L4, []AnyFactor{two("A"), two("B"), two("C"), two("D")}},
		{"duplicate name", L4, []AnyFactor{two("A"), two("A")}},
		{"same factor twice", L4, func() []AnyFactor { f := two("A"); return []AnyFactor{f, f} }()},
		{"empty name", L4, []AnyFactor{two("")}},
		{"single level", L4, []AnyFactor{NewFactor("A", 1)}},
		{"no levels", L4, []AnyFactor{NewFactor[int]("A")}},
		{"nil factor", L4, []AnyFactor{two("A"), nil}},
	}
	for _, c := range cases {
		if _, err := NewDesign(c.array, c.factors...); err == nil {
			t.Errorf("%s: accepted", c.name)
		}
	}
}

// A factor is a value: the same factors can take part in a screening
// design and a larger follow-up, and Of and Best answer for whichever
// design the run or analysis came from.
func TestFactorReuseAcrossDesigns(t *testing.T) {
	a := NewFactor("A", 10, 20)
	b := NewFactor("B", "x", "y")
	small, err := NewDesign(L4, a, b)
	if err != nil {
		t.Fatal(err)
	}
	large, err := NewDesign(L8.Select(0, 1, 3), b, a, NewFactor("C", 1, 2))
	if err != nil {
		t.Fatal(err)
	}
	for i, run := range small.Runs() {
		if a.Of(run) != a.Level(L4.Level(i, 0)) || b.Of(run) != b.Level(L4.Level(i, 1)) {
			t.Errorf("small run %d", i)
		}
	}
	for i, run := range large.Runs() {
		if b.Of(run) != b.Level(L8.Level(i, 0)) || a.Of(run) != a.Level(L8.Level(i, 1)) {
			t.Errorf("large run %d", i)
		}
	}
	exp := NewExperiment(large)
	for _, run := range large.Runs() {
		exp.Observe(run, float64(a.Of(run)))
	}
	an, err := exp.Analyze(SmallerTheBetter)
	if err != nil {
		t.Fatal(err)
	}
	if a.Best(an) != 10 {
		t.Errorf("Best in the larger design: %d", a.Best(an))
	}
}

func TestFactorOf(t *testing.T) {
	workers := NewFactor("workers", 1, 20)
	algo := NewFactor("algorithm", quick, radix)
	design, err := NewDesign(L4, workers, algo)
	if err != nil {
		t.Fatal(err)
	}
	runs := design.Runs()
	if len(runs) != 4 {
		t.Fatalf("%d runs", len(runs))
	}
	for i, run := range runs {
		if run != design.Run(i) || run.Row() != i {
			t.Errorf("run %d identity", i)
		}
		if got, want := workers.Of(run), workers.Level(L4.Level(i, 0)); got != want {
			t.Errorf("run %d: workers %d, want %d", i, got, want)
		}
		if got, want := algo.Of(run), algo.Level(L4.Level(i, 1)); got != want {
			t.Errorf("run %d: algorithm %v, want %v", i, got, want)
		}
		if run.Level(1) != L4.Level(i, 1) {
			t.Errorf("run %d: Level(1)", i)
		}
	}
	if s := runs[1].String(); s != "run 1: workers=1 algorithm=radix" {
		t.Errorf("Run.String: %q", s)
	}
	if s := (Run{}).String(); s != "run ?" {
		t.Errorf("zero Run.String: %q", s)
	}
	if s := algo.String(); s != "algorithm[quick radix]" {
		t.Errorf("Factor.String: %q", s)
	}
	levels := workers.Levels()
	levels[0] = 99
	if workers.Level(0) != 1 {
		t.Error("Levels returned the internal slice")
	}
	if names := design.FactorNames(); len(names) != 2 || names[0] != "workers" || names[1] != "algorithm" {
		t.Errorf("FactorNames: %v", names)
	}
	if design.Array().Name() != "L4" {
		t.Error("Array")
	}
	want := "Design: L4, 4 runs\n" +
		"RUN  workers  algorithm\n" +
		"0    1        quick\n" +
		"1    1        radix\n" +
		"2    20       quick\n" +
		"3    20       radix\n"
	if got := design.String(); got != want {
		t.Errorf("Design.String:\n%s\nwant:\n%s", got, want)
	}
}

func TestFactorPanics(t *testing.T) {
	loose := NewFactor("loose", 1, 2)
	a := NewFactor("a", 1, 2)
	b := NewFactor("b", 1, 2)
	d1, err := NewDesign(L4, a)
	if err != nil {
		t.Fatal(err)
	}
	d2, err := NewDesign(L4, b)
	if err != nil {
		t.Fatal(err)
	}
	mustPanic(t, "factor not in design", func() { loose.Of(d1.Run(0)) })
	mustPanic(t, "run from another design", func() { a.Of(d2.Run(0)) })
	mustPanic(t, "zero run", func() { a.Of(Run{}) })
	mustPanic(t, "zero run level", func() { (Run{}).Level(0) })
	mustPanic(t, "run out of range", func() { d1.Run(4) })
	mustPanic(t, "analysis without a design", func() {
		an, err := AnalyzeSNR(L4.Select(0), []float64{1, 2, 3, 4})
		if err != nil {
			t.Fatal(err)
		}
		a.Best(an)
	})
	mustPanic(t, "nil analysis", func() { a.Best(nil) })
	mustPanic(t, "analysis from another design", func() {
		exp := NewExperiment(d2)
		for _, run := range d2.Runs() {
			exp.Observe(run, 1, 2)
		}
		an, err := exp.Analyze(SmallerTheBetter)
		if err != nil {
			t.Fatal(err)
		}
		a.Best(an)
	})
}
