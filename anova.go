package taguchi

import (
	"fmt"
	"math"
	"slices"
	"strconv"
)

// Analysis is the result of a Taguchi analysis: the SNR of every run, the
// main effect of every factor, and the analysis of variance of the run
// SNRs. Quantities that are not defined, such as F when there is no error
// term, are NaN.
type Analysis struct {
	// Runs holds the pooled SNR of every run, in array order.
	Runs []RunResult
	// Mean is the grand mean of the run SNRs.
	Mean float64
	// TotalSS and TotalDF are the total sum of squares of the run SNRs
	// about the mean and its degrees of freedom, the number of runs minus one.
	TotalSS float64
	TotalDF int
	// Effects holds one entry per factor, in column order.
	Effects []Effect
	// Error is the residual: the variation not explained by the factors,
	// plus any factors that have been pooled into it.
	Error ErrorTerm
	// SNR is the ratio the run values were computed with. It is the zero
	// SNR when the analysis was produced by AnalyzeSNR.
	SNR SNR

	array  OrthogonalArray
	design *Design
}

// RunResult is the pooled SNR of one run.
type RunResult struct {
	// Row is the zero-based row of the run in the array.
	Row int
	// N is the number of observations pooled into the SNR. It is zero when
	// the analysis was produced by AnalyzeSNR rather than an Experiment.
	N int
	// SNR is the run's signal-to-noise ratio in decibels.
	SNR float64
}

// Effect is the main effect and ANOVA row of one factor.
type Effect struct {
	// Factor is the factor's name.
	Factor string
	// Levels holds a label for each level, in level order.
	Levels []string
	// Means holds the mean SNR of the runs at each level, in level order.
	// Their differences are the factor's main effect.
	Means []float64
	// Best is the index of the level with the highest mean SNR. Ties go to
	// the lowest index.
	Best int
	// SS, DF and MS are the factor's sum of squares, degrees of freedom and
	// mean square.
	SS float64
	DF int
	MS float64
	// F is the ratio MS / Error.MS and P the probability of an F at least
	// this large if the factor had no effect. Both are NaN when there is no
	// error term or when the factor is pooled.
	F float64
	P float64
	// Contribution is the factor's share of TotalSS in percent, NaN when
	// pooled or when TotalSS is zero.
	Contribution float64
	// Pooled reports that the factor's SS and DF have been added to the
	// error term by Pool.
	Pooled bool

	column int // column of the array; Effects may be reordered freely
}

// ErrorTerm is the error row of the ANOVA table.
type ErrorTerm struct {
	SS float64
	DF int
	// MS is SS / DF, NaN when DF is zero.
	MS float64
	// Contribution is the error's share of TotalSS in percent.
	Contribution float64
}

// AnalyzeSNR decomposes one SNR value per run of array, treating every
// column of array as a factor. It is the arithmetic behind
// Experiment.Analyze for callers who compute SNR values themselves; use
// OrthogonalArray.Select to pass exactly the columns that carry factors.
// Effects are named by column and levels by index.
//
// For every column the mean SNR at each level is computed and the column's
// sum of squares is the level count times the squared deviation of each
// level mean from the grand mean, summed over levels. The error sum of
// squares is what remains of the total, with the remaining degrees of
// freedom.
func AnalyzeSNR(array OrthogonalArray, snr []float64) (*Analysis, error) {
	n := array.Runs()
	if n == 0 {
		return nil, fmt.Errorf("taguchi: AnalyzeSNR: empty array")
	}
	factors := array.Columns()
	if len(snr) != n {
		return nil, fmt.Errorf("taguchi: AnalyzeSNR: %d SNR values for %d runs", len(snr), n)
	}
	for i, v := range snr {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return nil, fmt.Errorf("%w: run %d has SNR %v", ErrUndefinedSNR, i, v)
		}
	}

	a := &Analysis{
		Runs:    make([]RunResult, n),
		TotalDF: n - 1,
		Effects: make([]Effect, factors),
		array:   array,
	}
	for i, v := range snr {
		a.Runs[i] = RunResult{Row: i, SNR: v}
		a.Mean += v
	}
	a.Mean /= float64(n)
	for _, v := range snr {
		a.TotalSS += (v - a.Mean) * (v - a.Mean)
	}

	explainedSS, explainedDF := 0.0, 0
	for j := 0; j < factors; j++ {
		s := array.Levels(j)
		sums := make([]float64, s)
		counts := make([]int, s)
		for i := 0; i < n; i++ {
			l := array.Level(i, j)
			sums[l] += snr[i]
			counts[l]++
		}
		e := Effect{
			Factor: "column " + strconv.Itoa(j),
			Levels: make([]string, s),
			Means:  make([]float64, s),
			DF:     s - 1,
			column: j,
		}
		for l := 0; l < s; l++ {
			e.Levels[l] = strconv.Itoa(l)
			e.Means[l] = sums[l] / float64(counts[l])
			e.SS += float64(counts[l]) * (e.Means[l] - a.Mean) * (e.Means[l] - a.Mean)
			if e.Means[l] > e.Means[e.Best] {
				e.Best = l
			}
		}
		a.Effects[j] = e
		explainedSS += e.SS
		explainedDF += e.DF
	}

	// For an orthogonal array the factor sums of squares are projections
	// onto mutually orthogonal subspaces, so the residual is non-negative
	// up to rounding; clamp the rounding.
	a.Error.SS = math.Max(0, a.TotalSS-explainedSS)
	a.Error.DF = a.TotalDF - explainedDF
	a.recompute()
	return a, nil
}

// recompute derives mean squares, F, P and contributions from the sums of
// squares and degrees of freedom, honouring the current pooling.
func (a *Analysis) recompute() {
	nan := math.NaN()
	if a.Error.DF > 0 {
		a.Error.MS = a.Error.SS / float64(a.Error.DF)
	} else {
		a.Error.MS = nan
	}
	a.Error.Contribution = percent(a.Error.SS, a.TotalSS)
	for i := range a.Effects {
		e := &a.Effects[i]
		e.MS = e.SS / float64(e.DF)
		if e.Pooled {
			e.F, e.P, e.Contribution = nan, nan, nan
			continue
		}
		e.Contribution = percent(e.SS, a.TotalSS)
		if a.Error.DF > 0 {
			e.F = e.MS / a.Error.MS
			e.P = fPValue(e.F, e.DF, a.Error.DF)
		} else {
			e.F, e.P = nan, nan
		}
	}
}

func percent(part, total float64) float64 {
	if total == 0 {
		return math.NaN()
	}
	return 100 * part / total
}

// Saturated reports whether the factors use every degree of freedom, leaving
// no error term. F and P are then NaN; use Pool to obtain an error estimate
// from the least significant factors.
func (a *Analysis) Saturated() bool { return a.Error.DF == 0 }

// Balanced reports whether every run pooled the same number of observations.
// Runs with fewer observations have less precise SNR values, so an
// unbalanced experiment weights its runs unequally. It is true for an
// analysis produced by AnalyzeSNR, which carries no observation counts.
func (a *Analysis) Balanced() bool {
	if len(a.Runs) == 0 {
		return true
	}
	for _, r := range a.Runs {
		if r.N != a.Runs[0].N {
			return false
		}
	}
	return true
}

// Effect returns a copy of the effect of the named factor.
func (a *Analysis) Effect(factor string) (Effect, bool) {
	i := a.effectIndex(factor)
	if i < 0 {
		return Effect{}, false
	}
	e := a.Effects[i]
	e.Levels = slices.Clone(e.Levels)
	e.Means = slices.Clone(e.Means)
	return e, true
}

func (a *Analysis) effectIndex(factor string) int {
	return slices.IndexFunc(a.Effects, func(e Effect) bool { return e.Factor == factor })
}

// Pool returns a copy of the analysis in which the named factors are pooled
// into the error term: their sums of squares and degrees of freedom are
// added to the error, and F, P and contributions are recomputed for the
// remaining factors. Pooling the factors with the smallest sums of squares
// is the usual way to obtain an error estimate from a saturated design.
func (a *Analysis) Pool(factors ...string) (*Analysis, error) {
	b := a.clone()
	for _, name := range factors {
		i := b.effectIndex(name)
		if i < 0 {
			return nil, fmt.Errorf("taguchi: Pool: no factor named %q", name)
		}
		e := &b.Effects[i]
		if e.Pooled {
			return nil, fmt.Errorf("taguchi: Pool: factor %q is already pooled", name)
		}
		e.Pooled = true
		b.Error.SS += e.SS
		b.Error.DF += e.DF
	}
	b.recompute()
	return b, nil
}

// PredictedSNR returns the SNR predicted for the run that sets every
// unpooled factor to its best level, under the additive model:
//
//	Mean + sum over factors of (Means[Best] - Mean)
//
// Compare it with a confirmation run at those settings.
func (a *Analysis) PredictedSNR() float64 {
	snr := a.Mean
	for _, e := range a.Effects {
		if !e.Pooled {
			snr += e.Means[e.Best] - a.Mean
		}
	}
	return snr
}

func (a *Analysis) clone() *Analysis {
	b := *a
	b.Runs = slices.Clone(a.Runs)
	b.Effects = make([]Effect, len(a.Effects))
	for i, e := range a.Effects {
		e.Levels = slices.Clone(e.Levels)
		e.Means = slices.Clone(e.Means)
		b.Effects[i] = e
	}
	return &b
}
