package taguchi

import (
	"fmt"
	"slices"
	"strings"
)

// Factor is a control factor: a name and an ordered list of levels of any
// type T. A factor is a plain value with an identity: NewDesign records
// which factors occupy its columns, and Of and Best look the factor up in
// the design that a Run or an Analysis came from. The same factor can
// therefore take part in any number of designs, for example a screening
// design and a larger follow-up.
type Factor[T any] struct {
	name   string
	levels []T
	labels []string
}

// NewFactor declares a control factor. A factor needs at least two levels;
// the check is made by NewDesign, which is the only place a factor can go.
func NewFactor[T any](name string, levels ...T) *Factor[T] {
	f := &Factor[T]{
		name:   name,
		levels: slices.Clone(levels),
		labels: make([]string, len(levels)),
	}
	for i, l := range levels {
		f.labels[i] = fmt.Sprint(l)
	}
	return f
}

// Name returns the factor's name.
func (f *Factor[T]) Name() string { return f.name }

// Levels returns a copy of the factor's levels.
func (f *Factor[T]) Levels() []T { return slices.Clone(f.levels) }

// Level returns the level with the given zero-based index.
func (f *Factor[T]) Level(i int) T { return f.levels[i] }

// Of returns the factor's level in run. It panics if run is the zero Run or
// comes from a design that does not contain the factor; both are
// programming errors, not data conditions.
func (f *Factor[T]) Of(run Run) T {
	column := f.column(run.design)
	return f.levels[run.design.array.Level(run.row, column)]
}

// Best returns the factor's level with the highest mean SNR in a. It panics
// if a is nil or was not produced from a design that contains the factor.
func (f *Factor[T]) Best(a *Analysis) T {
	if a == nil {
		panic(fmt.Sprintf("taguchi: factor %q: Best of a nil analysis", f.name))
	}
	column := f.column(a.design)
	for _, e := range a.Effects {
		if e.column == column {
			return f.levels[e.Best]
		}
	}
	panic(fmt.Sprintf("taguchi: analysis has no effect for factor %q", f.name))
}

// String renders the factor as name[level level ...].
func (f *Factor[T]) String() string {
	return f.name + "[" + strings.Join(f.labels, " ") + "]"
}

// column returns the column the factor occupies in d.
func (f *Factor[T]) column(d *Design) int {
	if d == nil {
		panic(fmt.Sprintf("taguchi: factor %q: value does not come from a design", f.name))
	}
	if j := d.columnOf(f); j >= 0 {
		return j
	}
	panic(fmt.Sprintf("taguchi: factor %q is not in this design", f.name))
}

func (f *Factor[T]) numLevels() int         { return len(f.levels) }
func (f *Factor[T]) label(level int) string { return f.labels[level] }

// AnyFactor is the view of a *Factor[T] that does not depend on T. It is
// what NewDesign accepts, so factors with different level types can share a
// design. Only *Factor[T] implements it.
type AnyFactor interface {
	Name() string
	numLevels() int
	label(level int) string
}
