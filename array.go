package taguchi

import (
	"fmt"
	"strconv"
	"strings"
)

// OrthogonalArray is a table of N runs by k columns in which every column is
// balanced and every pair of columns is balanced: each combination of the
// levels of two columns occurs equally often. This strength-2 property is
// what lets the main effect of every factor be estimated independently of
// the other factors.
//
// An OrthogonalArray is immutable and safe to share. Level indices returned
// by Level are zero-based so that they can index slices directly.
type OrthogonalArray struct {
	name   string
	levels []int // number of levels of each column
	cells  []int // row-major, zero-based level indices
}

// NewOrthogonalArray validates rows and returns them as an OrthogonalArray.
// rows is given in the form found in reference tables: one entry per column
// in every row, with levels numbered from 1. Columns may have different
// numbers of levels, as in L18. The returned error wraps ErrNotOrthogonal
// when a column or a pair of columns is unbalanced.
func NewOrthogonalArray(name string, rows [][]int) (OrthogonalArray, error) {
	n := len(rows)
	if n < 2 {
		return OrthogonalArray{}, fmt.Errorf("taguchi: array %s: need at least 2 runs, got %d", name, n)
	}
	k := len(rows[0])
	if k == 0 {
		return OrthogonalArray{}, fmt.Errorf("taguchi: array %s: rows have no columns", name)
	}
	for i, row := range rows {
		if len(row) != k {
			return OrthogonalArray{}, fmt.Errorf("taguchi: array %s: rows[%d] has %d columns, want %d", name, i, len(row), k)
		}
	}

	a := OrthogonalArray{name: name, levels: make([]int, k), cells: make([]int, n*k)}
	for j := 0; j < k; j++ {
		s := 0
		for i := 0; i < n; i++ {
			v := rows[i][j]
			if v < 1 {
				return OrthogonalArray{}, fmt.Errorf("taguchi: array %s: rows[%d][%d] is %d, levels are numbered from 1", name, i, j, v)
			}
			if v > s {
				s = v
			}
			a.cells[i*k+j] = v - 1
		}
		if s < 2 {
			return OrthogonalArray{}, fmt.Errorf("taguchi: array %s: column %d has a single level", name, j)
		}
		a.levels[j] = s
	}

	for j := 0; j < k; j++ {
		s := a.levels[j]
		if n%s != 0 {
			return OrthogonalArray{}, fmt.Errorf("%w: %s column %d: %d runs cannot balance %d levels", ErrNotOrthogonal, name, j, n, s)
		}
		counts := make([]int, s)
		for i := 0; i < n; i++ {
			counts[a.cells[i*k+j]]++
		}
		for l, c := range counts {
			if c != n/s {
				return OrthogonalArray{}, fmt.Errorf("%w: %s column %d: level %d occurs %d times, want %d", ErrNotOrthogonal, name, j, l+1, c, n/s)
			}
		}
	}

	for x := 0; x < k; x++ {
		for y := x + 1; y < k; y++ {
			sx, sy := a.levels[x], a.levels[y]
			if n%(sx*sy) != 0 {
				return OrthogonalArray{}, fmt.Errorf("%w: %s columns %d and %d: %d runs cannot balance %d level pairs", ErrNotOrthogonal, name, x, y, n, sx*sy)
			}
			want := n / (sx * sy)
			counts := make([]int, sx*sy)
			for i := 0; i < n; i++ {
				counts[a.cells[i*k+x]*sy+a.cells[i*k+y]]++
			}
			for c, got := range counts {
				if got != want {
					return OrthogonalArray{}, fmt.Errorf("%w: %s columns %d and %d: levels (%d,%d) occur %d times, want %d",
						ErrNotOrthogonal, name, x, y, c/sy+1, c%sy+1, got, want)
				}
			}
		}
	}
	return a, nil
}

// Name returns the name the array was created with, such as "L8".
func (a OrthogonalArray) Name() string { return a.name }

// Runs returns the number of runs (rows) of the array.
func (a OrthogonalArray) Runs() int {
	if len(a.levels) == 0 {
		return 0
	}
	return len(a.cells) / len(a.levels)
}

// Columns returns the number of columns of the array.
func (a OrthogonalArray) Columns() int { return len(a.levels) }

// Levels returns the number of levels of a column.
func (a OrthogonalArray) Levels(column int) int {
	a.checkColumn(column)
	return a.levels[column]
}

// Level returns the zero-based level index in the given run and column.
func (a OrthogonalArray) Level(run, column int) int {
	a.checkColumn(column)
	if run < 0 || run >= a.Runs() {
		panic(fmt.Sprintf("taguchi: run %d out of range [0,%d)", run, a.Runs()))
	}
	return a.cells[run*len(a.levels)+column]
}

// Select returns the array made of the given columns, in the given order.
// Every subset of the columns of an orthogonal array is itself orthogonal,
// so the result needs no validation. Use it to place factors on specific
// columns, for example to keep the columns of L8 or L16 that carry
// two-factor interactions free of factors:
//
//	design, err := taguchi.NewDesign(taguchi.L8.Select(0, 1, 3), a, b, c)
//
// It panics if a column is out of range or selected twice.
func (a OrthogonalArray) Select(columns ...int) OrthogonalArray {
	if len(columns) == 0 {
		panic("taguchi: Select: no columns")
	}
	seen := make(map[int]bool, len(columns))
	for _, c := range columns {
		a.checkColumn(c)
		if seen[c] {
			panic(fmt.Sprintf("taguchi: Select: column %d selected twice", c))
		}
		seen[c] = true
	}
	n, k := a.Runs(), len(columns)
	b := OrthogonalArray{
		name:   fmt.Sprintf("%s%v", a.name, columns),
		levels: make([]int, k),
		cells:  make([]int, n*k),
	}
	for j, c := range columns {
		b.levels[j] = a.levels[c]
		for i := 0; i < n; i++ {
			b.cells[i*k+j] = a.cells[i*len(a.levels)+c]
		}
	}
	return b
}

func (a OrthogonalArray) checkColumn(column int) {
	if column < 0 || column >= len(a.levels) {
		panic(fmt.Sprintf("taguchi: column %d out of range [0,%d)", column, len(a.levels)))
	}
}

// String renders the array as it appears in reference tables, with levels
// numbered from 1.
func (a OrthogonalArray) String() string {
	var b strings.Builder
	n, k := a.Runs(), a.Columns()
	fmt.Fprintf(&b, "%s (%d runs, %d columns)\n", a.name, n, k)
	for i := 0; i < n; i++ {
		for j := 0; j < k; j++ {
			if j > 0 {
				b.WriteByte(' ')
			}
			b.WriteString(strconv.Itoa(a.cells[i*k+j] + 1))
		}
		b.WriteByte('\n')
	}
	return b.String()
}

func mustArray(name string, rows [][]int) OrthogonalArray {
	a, err := NewOrthogonalArray(name, rows)
	if err != nil {
		panic(err)
	}
	return a
}

// Standard orthogonal arrays, in the column order used by the Taguchi
// reference tables. L(N) has N runs. The 2-level arrays L4, L8 and L16 are
// generated by the same construction and so agree with the interaction
// tables and linear graphs printed for them; L12 has no such columns.
// L18 has one 2-level column followed by seven 3-level columns.
//
// Any other array can be supplied through NewOrthogonalArray.
var (
	// L4 accommodates up to three 2-level factors.
	L4 = mustArray("L4", [][]int{
		{1, 1, 1},
		{1, 2, 2},
		{2, 1, 2},
		{2, 2, 1},
	})

	// L8 accommodates up to seven 2-level factors.
	L8 = mustArray("L8", [][]int{
		{1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 2, 2, 2, 2},
		{1, 2, 2, 1, 1, 2, 2},
		{1, 2, 2, 2, 2, 1, 1},
		{2, 1, 2, 1, 2, 1, 2},
		{2, 1, 2, 2, 1, 2, 1},
		{2, 2, 1, 1, 2, 2, 1},
		{2, 2, 1, 2, 1, 1, 2},
	})

	// L9 accommodates up to four 3-level factors.
	L9 = mustArray("L9", [][]int{
		{1, 1, 1, 1},
		{1, 2, 2, 2},
		{1, 3, 3, 3},
		{2, 1, 2, 3},
		{2, 2, 3, 1},
		{2, 3, 1, 2},
		{3, 1, 3, 2},
		{3, 2, 1, 3},
		{3, 3, 2, 1},
	})

	// L12 accommodates up to eleven 2-level factors. Interactions are spread
	// evenly over all columns rather than confounded with any single one.
	L12 = mustArray("L12", [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2},
		{1, 1, 2, 2, 2, 1, 1, 1, 2, 2, 2},
		{1, 2, 1, 2, 2, 1, 2, 2, 1, 1, 2},
		{1, 2, 2, 1, 2, 2, 1, 2, 1, 2, 1},
		{1, 2, 2, 2, 1, 2, 2, 1, 2, 1, 1},
		{2, 1, 2, 2, 1, 1, 2, 2, 1, 2, 1},
		{2, 1, 2, 1, 2, 2, 2, 1, 1, 1, 2},
		{2, 1, 1, 2, 2, 2, 1, 2, 2, 1, 1},
		{2, 2, 2, 1, 1, 1, 1, 2, 2, 1, 2},
		{2, 2, 1, 2, 1, 2, 1, 1, 1, 2, 2},
		{2, 2, 1, 1, 2, 1, 2, 1, 2, 2, 1},
	})

	// L16 accommodates up to fifteen 2-level factors.
	L16 = mustArray("L16", [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 1, 1, 1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2},
		{1, 1, 1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2},
		{1, 1, 1, 2, 2, 2, 2, 2, 2, 2, 2, 1, 1, 1, 1},
		{1, 2, 2, 1, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2, 2},
		{1, 2, 2, 1, 1, 2, 2, 2, 2, 1, 1, 2, 2, 1, 1},
		{1, 2, 2, 2, 2, 1, 1, 1, 1, 2, 2, 2, 2, 1, 1},
		{1, 2, 2, 2, 2, 1, 1, 2, 2, 1, 1, 1, 1, 2, 2},
		{2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2, 1, 2},
		{2, 1, 2, 1, 2, 1, 2, 2, 1, 2, 1, 2, 1, 2, 1},
		{2, 1, 2, 2, 1, 2, 1, 1, 2, 1, 2, 2, 1, 2, 1},
		{2, 1, 2, 2, 1, 2, 1, 2, 1, 2, 1, 1, 2, 1, 2},
		{2, 2, 1, 1, 2, 2, 1, 1, 2, 2, 1, 1, 2, 2, 1},
		{2, 2, 1, 1, 2, 2, 1, 2, 1, 1, 2, 2, 1, 1, 2},
		{2, 2, 1, 2, 1, 1, 2, 1, 2, 2, 1, 2, 1, 1, 2},
		{2, 2, 1, 2, 1, 1, 2, 2, 1, 1, 2, 1, 2, 2, 1},
	})

	// L18 accommodates one 2-level factor and up to seven 3-level factors.
	L18 = mustArray("L18", [][]int{
		{1, 1, 1, 1, 1, 1, 1, 1},
		{1, 1, 2, 2, 2, 2, 2, 2},
		{1, 1, 3, 3, 3, 3, 3, 3},
		{1, 2, 1, 1, 2, 2, 3, 3},
		{1, 2, 2, 2, 3, 3, 1, 1},
		{1, 2, 3, 3, 1, 1, 2, 2},
		{1, 3, 1, 2, 1, 3, 2, 3},
		{1, 3, 2, 3, 2, 1, 3, 1},
		{1, 3, 3, 1, 3, 2, 1, 2},
		{2, 1, 1, 3, 3, 2, 2, 1},
		{2, 1, 2, 1, 1, 3, 3, 2},
		{2, 1, 3, 2, 2, 1, 1, 3},
		{2, 2, 1, 2, 3, 1, 3, 2},
		{2, 2, 2, 3, 1, 2, 1, 3},
		{2, 2, 3, 1, 2, 3, 2, 1},
		{2, 3, 1, 3, 2, 3, 1, 2},
		{2, 3, 2, 1, 3, 1, 2, 3},
		{2, 3, 3, 2, 1, 2, 3, 1},
	})
)
