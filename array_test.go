package taguchi

import (
	"errors"
	"strings"
	"testing"
)

// strength2 is an independent check of the orthogonal array property, kept
// separate from NewOrthogonalArray so that the two cannot share a bug.
// rows uses the 1-based reference form.
func strength2(rows [][]int) bool {
	k := len(rows[0])
	for a := 0; a < k; a++ {
		for b := a + 1; b < k; b++ {
			counts := map[[2]int]int{}
			for _, r := range rows {
				counts[[2]int{r[a], r[b]}]++
			}
			want := -1
			for _, c := range counts {
				if want == -1 {
					want = c
				} else if c != want {
					return false
				}
			}
			levelsA, levelsB := map[int]bool{}, map[int]bool{}
			for _, r := range rows {
				levelsA[r[a]] = true
				levelsB[r[b]] = true
			}
			if len(counts) != len(levelsA)*len(levelsB) {
				return false // some combination never occurs
			}
		}
	}
	return true
}

func table(a OrthogonalArray) [][]int {
	rows := make([][]int, a.Runs())
	for i := range rows {
		rows[i] = make([]int, a.Columns())
		for j := range rows[i] {
			rows[i][j] = a.Level(i, j) + 1
		}
	}
	return rows
}

// twoLevelArray builds L(2^bits) by the standard construction: column c
// (1-based) is the parity of the row index masked with c. This is the
// column order of the reference tables for L4, L8 and L16.
func twoLevelArray(bits int) [][]int {
	n := 1 << bits
	rows := make([][]int, n)
	for r := 0; r < n; r++ {
		rows[r] = make([]int, n-1)
		for c := 1; c < n; c++ {
			v := 0
			for k := 0; k < bits; k++ {
				if c&(1<<k) != 0 {
					v ^= (r >> (bits - 1 - k)) & 1
				}
			}
			rows[r][c-1] = v + 1
		}
	}
	return rows
}

func TestStandardArraysAreOrthogonal(t *testing.T) {
	cases := []struct {
		array  OrthogonalArray
		runs   int
		levels []int
	}{
		{L4, 4, []int{2, 2, 2}},
		{L8, 8, []int{2, 2, 2, 2, 2, 2, 2}},
		{L9, 9, []int{3, 3, 3, 3}},
		{L12, 12, []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}},
		{L16, 16, []int{2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2, 2}},
		{L18, 18, []int{2, 3, 3, 3, 3, 3, 3, 3}},
	}
	for _, c := range cases {
		a := c.array
		if a.Runs() != c.runs || a.Columns() != len(c.levels) {
			t.Errorf("%s: %dx%d, want %dx%d", a.Name(), a.Runs(), a.Columns(), c.runs, len(c.levels))
		}
		for j, s := range c.levels {
			if a.Levels(j) != s {
				t.Errorf("%s column %d: %d levels, want %d", a.Name(), j, a.Levels(j), s)
			}
		}
		if !strength2(table(a)) {
			t.Errorf("%s is not a strength-2 orthogonal array", a.Name())
		}
	}
}

func TestTwoLevelArraysMatchConstruction(t *testing.T) {
	for bits, a := range map[int]OrthogonalArray{2: L4, 3: L8, 4: L16} {
		want := twoLevelArray(bits)
		got := table(a)
		for i := range want {
			for j := range want[i] {
				if got[i][j] != want[i][j] {
					t.Fatalf("%s row %d: got %v, want %v", a.Name(), i, got[i], want[i])
				}
			}
		}
	}
}

func TestNewOrthogonalArrayRejects(t *testing.T) {
	cases := []struct {
		name       string
		rows       [][]int
		orthogonal bool // expect ErrNotOrthogonal rather than a shape error
	}{
		{"one row", [][]int{{1, 2}}, false},
		{"no columns", [][]int{{}, {}}, false},
		{"ragged", [][]int{{1, 1}, {2}}, false},
		{"level zero", [][]int{{0, 1}, {1, 2}}, false},
		{"single level column", [][]int{{1, 1}, {1, 2}}, false},
		{"unbalanced column", [][]int{{1, 1}, {1, 2}, {1, 1}, {2, 2}}, true},
		{"confounded columns", [][]int{{1, 1}, {1, 1}, {2, 2}, {2, 2}}, true},
		{"too few runs for pairs", [][]int{{1, 1}, {2, 2}}, true},
	}
	for _, c := range cases {
		_, err := NewOrthogonalArray(c.name, c.rows)
		if err == nil {
			t.Errorf("%s: accepted", c.name)
			continue
		}
		if errors.Is(err, ErrNotOrthogonal) != c.orthogonal {
			t.Errorf("%s: errors.Is(ErrNotOrthogonal) = %v, want %v: %v", c.name, !c.orthogonal, c.orthogonal, err)
		}
	}
}

func TestNewOrthogonalArrayAcceptsMixedLevels(t *testing.T) {
	a, err := NewOrthogonalArray("copy", table(L18))
	if err != nil {
		t.Fatal(err)
	}
	if a.Levels(0) != 2 || a.Levels(1) != 3 {
		t.Errorf("levels: %d, %d", a.Levels(0), a.Levels(1))
	}
}

func TestArrayAccessors(t *testing.T) {
	if L18.Level(3, 1) != 1 { // row 4 of the table is 1 2 1 1 2 2 3 3
		t.Errorf("L18.Level(3,1) = %d, want 1", L18.Level(3, 1))
	}
	if !strings.HasPrefix(L18.String(), "L18 (18 runs, 8 columns)\n1 1 1 1 1 1 1 1\n1 1 2 2 2 2 2 2\n") {
		t.Errorf("String:\n%s", L18.String())
	}
	var zero OrthogonalArray
	if zero.Runs() != 0 || zero.Columns() != 0 {
		t.Error("zero array has runs or columns")
	}
	mustPanic(t, "run out of range", func() { L4.Level(4, 0) })
	mustPanic(t, "column out of range", func() { L4.Level(0, 3) })
	mustPanic(t, "negative column", func() { L4.Levels(-1) })
}

func mustPanic(t *testing.T, name string, f func()) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("%s: no panic", name)
		}
	}()
	f()
}

func TestSelect(t *testing.T) {
	sub := L8.Select(0, 1, 3)
	if sub.Name() != "L8[0 1 3]" || sub.Runs() != 8 || sub.Columns() != 3 {
		t.Errorf("%s %dx%d", sub.Name(), sub.Runs(), sub.Columns())
	}
	for i := 0; i < 8; i++ {
		if sub.Level(i, 2) != L8.Level(i, 3) || sub.Level(i, 0) != L8.Level(i, 0) {
			t.Errorf("run %d: columns not copied", i)
		}
	}
	if !strength2(table(sub)) {
		t.Error("selection is not orthogonal")
	}
	if !strength2(table(L18.Select(7, 0, 3))) || L18.Select(7, 0).Levels(1) != 2 {
		t.Error("mixed-level selection")
	}
	mustPanic(t, "no columns", func() { L8.Select() })
	mustPanic(t, "out of range", func() { L8.Select(0, 7) })
	mustPanic(t, "duplicate", func() { L8.Select(1, 1) })
}

// In L8, column 2 is the interaction of columns 0 and 1. A factor placed
// there absorbs that interaction; placed on column 3 it does not, and the
// interaction shows up in the error term instead.
func TestSelectKeepsInteractionOutOfFactor(t *testing.T) {
	snr := make([]float64, 8)
	for i := range snr {
		a, b := L8.Level(i, 0), L8.Level(i, 1)
		snr[i] = -20 + 2*float64(a) - 3*float64(b) + 5*float64(a^b)
	}
	confounded, err := AnalyzeSNR(L8.Select(0, 1, 2), snr)
	if err != nil {
		t.Fatal(err)
	}
	clean, err := AnalyzeSNR(L8.Select(0, 1, 3), snr)
	if err != nil {
		t.Fatal(err)
	}
	if !near(confounded.Effects[2].SS, 50, 1e-9) || confounded.Error.SS > 1e-9 {
		t.Errorf("columns 0,1,2: C.SS %v error %v", confounded.Effects[2].SS, confounded.Error.SS)
	}
	if clean.Effects[2].SS > 1e-9 || !near(clean.Error.SS, 50, 1e-9) {
		t.Errorf("columns 0,1,3: C.SS %v error %v", clean.Effects[2].SS, clean.Error.SS)
	}
}
