package taguchi

import (
	"fmt"
	"strconv"
	"strings"
	"text/tabwriter"
)

// Design is an orthogonal array with a factor assigned to each of its first
// columns. It is immutable once created and safe to share; the same Design
// can back any number of Experiments.
type Design struct {
	array   OrthogonalArray
	factors []AnyFactor
}

// NewDesign assigns factors, in order, to the columns of array. Columns
// beyond the last factor stay unassigned and contribute to the error term
// of the analysis. Use OrthogonalArray.Select to choose which columns the
// factors occupy.
//
// It returns an error when a factor has fewer than two levels, when a
// factor's number of levels differs from the number of levels of its
// column, when names are empty or repeated, or when there are more factors
// than columns.
func NewDesign(array OrthogonalArray, factors ...AnyFactor) (*Design, error) {
	if array.Columns() == 0 {
		return nil, fmt.Errorf("taguchi: design: array has no columns")
	}
	if len(factors) == 0 {
		return nil, fmt.Errorf("taguchi: design: no factors")
	}
	if len(factors) > array.Columns() {
		return nil, fmt.Errorf("taguchi: design: %s has %d columns, cannot hold %d factors", array.Name(), array.Columns(), len(factors))
	}
	seen := make(map[string]bool, len(factors))
	for j, f := range factors {
		if f == nil {
			return nil, fmt.Errorf("taguchi: design: factor %d is nil", j)
		}
		name := f.Name()
		switch {
		case name == "":
			return nil, fmt.Errorf("taguchi: design: factor %d has no name", j)
		case seen[name]:
			return nil, fmt.Errorf("taguchi: design: factor %q appears twice", name)
		case f.numLevels() < 2:
			return nil, fmt.Errorf("taguchi: design: factor %q has %d levels, need at least 2", name, f.numLevels())
		case f.numLevels() != array.Levels(j):
			return nil, fmt.Errorf("taguchi: design: factor %q has %d levels but column %d of %s has %d",
				name, f.numLevels(), j, array.Name(), array.Levels(j))
		}
		seen[name] = true
	}
	return &Design{array: array, factors: append([]AnyFactor(nil), factors...)}, nil
}

// columnOf returns the column occupied by f, or -1 if f is not in the
// design. Factors are compared by identity.
func (d *Design) columnOf(f AnyFactor) int {
	for j, g := range d.factors {
		if g == f {
			return j
		}
	}
	return -1
}

// factorArray returns the columns of the array that carry factors.
func (d *Design) factorArray() OrthogonalArray {
	k := len(d.factors)
	if k == d.array.Columns() {
		return d.array
	}
	columns := make([]int, k)
	for j := range columns {
		columns[j] = j
	}
	return d.array.Select(columns...)
}

// Array returns the design's orthogonal array.
func (d *Design) Array() OrthogonalArray { return d.array }

// Runs returns every run of the design, in array order.
func (d *Design) Runs() []Run {
	runs := make([]Run, d.array.Runs())
	for i := range runs {
		runs[i] = Run{design: d, row: i}
	}
	return runs
}

// Run returns the run with the given zero-based row index.
func (d *Design) Run(row int) Run {
	if row < 0 || row >= d.array.Runs() {
		panic(fmt.Sprintf("taguchi: run %d out of range [0,%d)", row, d.array.Runs()))
	}
	return Run{design: d, row: row}
}

// FactorNames returns the factor names in column order.
func (d *Design) FactorNames() []string {
	names := make([]string, len(d.factors))
	for j, f := range d.factors {
		names[j] = f.Name()
	}
	return names
}

// String renders the plan: one line per run with the level of each factor.
func (d *Design) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "Design: %s, %d runs\n", d.array.Name(), d.array.Runs())
	tw := tabwriter.NewWriter(&b, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "RUN\t"+strings.Join(d.FactorNames(), "\t"))
	for i := 0; i < d.array.Runs(); i++ {
		fmt.Fprintln(tw, strconv.Itoa(i)+"\t"+strings.Join(d.labels(i), "\t"))
	}
	tw.Flush()
	return b.String()
}

// labels returns the level label of every factor in run i.
func (d *Design) labels(i int) []string {
	out := make([]string, len(d.factors))
	for j, f := range d.factors {
		out[j] = f.label(d.array.Level(i, j))
	}
	return out
}

// Run identifies one row of a Design. It is a small comparable value: two
// Runs are equal when they name the same row of the same design.
type Run struct {
	design *Design
	row    int
}

// Row returns the zero-based row index of the run in its design's array.
func (r Run) Row() int { return r.row }

// Level returns the zero-based level index of the given column in this run.
// Typed access through Factor.Of is usually more convenient.
func (r Run) Level(column int) int {
	if r.design == nil {
		panic("taguchi: zero Run has no design")
	}
	return r.design.array.Level(r.row, column)
}

// String renders the run as "run 3: name=level name=level ...".
func (r Run) String() string {
	if r.design == nil {
		return "run ?"
	}
	var b strings.Builder
	fmt.Fprintf(&b, "run %d:", r.row)
	for j, f := range r.design.factors {
		fmt.Fprintf(&b, " %s=%s", f.Name(), f.label(r.design.array.Level(r.row, j)))
	}
	return b.String()
}
