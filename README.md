# taguchi

A Go library for the Taguchi method of experimental design: find the settings
of the things you control that make a system perform best and most
consistently under the conditions you do not control, using far fewer runs
than trying every combination. No dependencies.

```bash
go get github.com/aleksicmarija/taguchi
```

## What it is for

You have a system with several knobs, each with a few candidate settings.
Trying every combination is expensive: three knobs with three settings each
is 27 combinations, seven two-level knobs is 128. The Taguchi method runs a
chosen subset, laid out by an orthogonal array, and from those runs
estimates the effect of every knob independently of the others. It answers
three questions:

- which settings are best
- how much each knob matters
- whether the evidence for that is solid or could be noise

"Best" in the Taguchi sense means good and stable. A setting that performs
well across the conditions you cannot control, such as input data, load or
hardware, beats one that is fastest in one case and slow in another. That
preference is built into the signal-to-noise ratio the method optimises.

The library provides the standard orthogonal arrays and validates custom
ones, assigns typed factors to them, collects observations, computes the
signal-to-noise ratios, main effects and analysis of variance with F and p
values, and prints a report.

## What everything means

In the order you meet them.

- **Control factor.** A knob whose setting you choose. It has a name and an
  ordered list of **levels**, the candidate settings. Levels can be values
  of any Go type.
- **Noise.** Conditions you do not control or choose not to fix. The method
  does not try to remove noise; it looks for settings that are insensitive
  to it. The library does not model noise factors: you execute each run
  under the noise conditions of your choice and record every observation.
- **Orthogonal array.** A table with one row per **run** and one **column**
  per factor slot. Every column is balanced, each level appearing equally
  often, and every pair of columns is balanced, each combination of two
  levels appearing equally often. That balance is what lets a handful of
  runs estimate every factor's effect without bias from the others. L8
  means eight runs. Built in: L4, L8, L9, L12, L16, L18.
- **Design.** An orthogonal array with a factor assigned to each of its
  leading columns. Its one law, checked once: each factor has exactly as
  many levels as its column.
- **Run.** One row of the design, one setting for every factor. You execute
  each run, usually several times under different noise conditions, and
  record every measured response as an **observation**.
- **Quality characteristic and SNR.** What you measure and which direction
  is good: smaller-the-better (latency, defects), larger-the-better
  (throughput, strength), nominal-the-best (hit a target). The
  **signal-to-noise ratio** turns all observations of one run into one
  number in decibels that rewards a good average and a small spread
  together. Higher is always better.
- **Main effect.** For one factor, the mean SNR of the runs at each of its
  levels. The differences between those means are the factor's effect; the
  level with the highest mean is its **best level**.
- **ANOVA.** Analysis of variance: splits the variation between the run
  SNRs into a share for each factor plus an unexplained remainder, the
  **error**, and tests each factor's share against it. Per factor: **SS**,
  the sum of squares, its share of variation; **DF**, degrees of freedom,
  the number of levels minus one; **MS**, SS divided by DF; **F**, MS
  divided by the error MS; **p**, the probability of an F at least this
  large if the factor did nothing; **contribution**, SS as a percentage of
  the total.
- **Saturated design.** The factors use every degree of freedom, so there is
  no error term and no F or p. The remedy is **pooling**: folding the
  weakest factors into the error.
- **Predicted SNR and confirmation run.** The SNR the additive model
  predicts at the best settings. Executing those settings once more and
  comparing tells you whether the model is adequate.

## How to use the API

### 1. Declare factors

```go
threads := taguchi.NewFactor("threads", 2, 4, 8)              // *Factor[int]
batch   := taguchi.NewFactor("batch", 10, 50, 250)            // *Factor[int]
comp    := taguchi.NewFactor("compression", None, Fast, Best) // *Factor[Compression]
```

Any type works as a level. Labels in reports come from `fmt.Sprint`, so a
type with a `String` method prints its name.

### 2. Build a design

```go
design, err := taguchi.NewDesign(taguchi.L9, threads, batch, comp)
```

Factors take the columns of the array in order. `NewDesign` returns an
error if a factor's level count differs from its column's, if names are
empty or repeated, or if there are more factors than columns. Columns
without a factor go to the error term. `fmt.Println(design)` prints the
plan as a table.

In the two-level arrays some columns carry the interaction of two others:
in L8, column 2 is the interaction of columns 0 and 1. A factor placed there
absorbs that interaction. `Select` picks the columns to use, and any subset
of an orthogonal array is still orthogonal:

```go
design, err := taguchi.NewDesign(taguchi.L8.Select(0, 1, 3), a, b, c)
```

Any array from a reference book can be used, in the printed form with
levels numbered from 1. It is validated on construction:

```go
l27, err := taguchi.NewOrthogonalArray("L27", rows) // errors.Is(err, taguchi.ErrNotOrthogonal)
```

### 3. Execute the runs and record observations

```go
exp := taguchi.NewExperiment(design)
for _, run := range design.Runs() {
	for _, payload := range payloads { // your noise conditions
		ms := measure(threads.Of(run), batch.Of(run), comp.Of(run), payload)
		exp.Observe(run, ms)
	}
}
```

`factor.Of(run)` returns the factor's level in that run, with the factor's
own type. Observe a run as many times as you like; all observations are
pooled. `Observe` is safe to call from several goroutines. `run.String()`
prints the run's settings for logging.

A factor is a plain value and can take part in several designs, such as a
screening design and a larger follow-up. `Of` looks it up in the design the
run came from.

### 4. Analyse

```go
analysis, err := exp.Analyze(taguchi.SmallerTheBetter)
if err != nil {
	// a run without observations, or data the SNR rejects; the run is named
}
fmt.Print(analysis)                // the report
best := threads.Best(analysis)     // 8, typed
e, _ := analysis.Effect("threads") // Means, Best, SS, DF, MS, F, P, Contribution
```

| Goal | Use |
|---|---|
| minimise the response | `taguchi.SmallerTheBetter` |
| maximise the response | `taguchi.LargerTheBetter` |
| sit on a target with the least variation | `taguchi.NominalTheBest` |
| minimise the distance from a target | `taguchi.NominalTheBestTarget(t)` |
| your own | `taguchi.SNR{Name: "...", Of: func(y []float64) (float64, error) { ... }}` |

`Analysis` holds `Runs` (SNR and observation count per run), `Mean`,
`TotalSS`, `TotalDF`, `Effects` (one per factor, in design order), `Error`
and `SNR`. Its methods are `Saturated`, `Pool`, `PredictedSNR`, `Balanced`,
`Effect`, `WriteTo` and `String`. `Effects` may be sorted freely; `Best` and
the report identify a factor by its column, not its position.

If the design is saturated, pool the weakest factors to obtain an error
estimate:

```go
if analysis.Saturated() {
	analysis, err = analysis.Pool("batch")
}
```

For your own pipeline, `AnalyzeSNR(array, snr)` does the arithmetic on one
SNR value per run, treating every column of the array as a factor.

### Conventions

- Every index is zero-based: runs, columns, levels, `Effect.Best`. The one
  exception is the input of `NewOrthogonalArray`, which takes the textbook
  form with levels numbered from 1.
- Data problems return errors wrapping `ErrNoObservations`, `ErrDomain`,
  `ErrUndefinedSNR` or `ErrNotOrthogonal`. Programming errors, such as
  passing a run from a different design, panic.
- Undefined quantities are `NaN` and print as `n/a`.

## How it is calculated

### Signal-to-noise ratio

For one run with observations y₁ … yₙ:

| SNR | Formula | Defined when |
|---|---|---|
| smaller-the-better | −10 log₁₀ ( (1/n) Σ yᵢ² ) | not all yᵢ = 0 |
| larger-the-better | −10 log₁₀ ( (1/n) Σ 1/yᵢ² ) | all yᵢ > 0 |
| nominal-the-best | 10 log₁₀ ( ȳ² / s² − 1/n ) | n ≥ 2, s² > 0, \|ȳ\| > s/√n |
| nominal-the-best, target T | −10 log₁₀ ( (1/n) Σ (yᵢ − T)² ) | not all yᵢ = T |

s² is the sample variance with n − 1 in the denominator. The
nominal-the-best form is Taguchi's own, 10 log₁₀((Sₘ − Vₑ)/(n Vₑ)) with
Sₘ = n ȳ² and Vₑ = s². Minitab and JMP omit the 1/n term, which changes the
value by under 0.05 dB once the mean is ten standard errors from zero. It
measures spread relative to the mean and ignores the target, which allows
two-step optimisation: first pick the levels that maximise the SNR, then
move the mean onto the target with a factor that only shifts the mean. The
target form penalises bias and spread together.

Why decibels: the smaller-the-better SNR is −20 log₁₀ of the
root-mean-square of y, so a gain of 6 dB means the RMS halved and 20 dB
means it fell tenfold. Squaring makes a few bad observations count more
than many good ones, which is what penalises a setting that is sometimes
slow.

Why the observations of a run are pooled before the log is taken: the log
is not linear, so the average of per-condition SNRs is not the SNR of all
the data. The library computes each run's SNR once, over every observation
of the run, and offers no other way to obtain one.

An SNR refuses input outside its domain instead of substituting a value,
because one infinite or fabricated number would corrupt every quantity
below it.

### ANOVA

Inputs: the N run SNRs η₁ … η_N and the array. With m their grand mean:

- total: SS_T = Σ (ηᵢ − m)², DF_T = N − 1
- factor A with s levels, each level appearing in N/s runs with mean
  m_{A,l}: SS_A = (N/s) Σ_l (m_{A,l} − m)², DF_A = s − 1
- error: SS_e = SS_T − Σ SS_factors, DF_e = DF_T − Σ DF_factors
- mean squares: MS = SS / DF
- F_A = MS_A / MS_e, and p_A = P( F(DF_A, DF_e) ≥ F_A ), the upper tail of
  the F distribution, evaluated with the regularised incomplete beta
  function
- contribution: 100 · SS_A / SS_T, and likewise for the error
- main effect of A: the level means m_{A,l}; best level: the largest
- predicted SNR at the best settings: m + Σ_A (m_{A,best} − m)

The balance of the array is what makes this valid. Every level of A meets
every level of B equally often, so the level means of A are unaffected by
B, and the factor sums of squares add up exactly to the part of SS_T they
explain. The error collects the unassigned columns and any interactions
between factors. When DF_e is 0 the design is saturated, MS_e is undefined,
and F and p are NaN. Pooling moves a factor's SS and DF into the error and
recomputes F and p for the rest. Sums of squares are computed in two passes
about the mean, and a residual that rounds slightly negative is clamped at
zero.

## What the result means

A report for three factors on L9, three observations per run,
smaller-the-better:

```
Taguchi analysis: L9, 9 runs, 3 factors, smaller-the-better

RUNS (SNR in dB, pooled over N observations)
run  threads  batch  compression  N  SNR
0    2        10     none         3  -30.427
1    2        50     fast         3  -30.913
2    2        250    best         3  -34.085
3    4        10     fast         3  -27.905
4    4        50     best         3  -30.897
5    4        250    none         3  -29.375
6    8        10     best         3  -28.626
7    8        50     none         3  -24.783
8    8        250    fast         3  -28.199
grand mean: -29.468 dB

MAIN EFFECTS (mean SNR per level, dB; * marks the best level)
factor       level  mean SNR
threads      2      -31.809
threads      4      -29.392
threads      8      -27.203   *
batch        10     -28.986
batch        50     -28.864   *
batch        250    -30.553
compression  none   -28.195   *
compression  fast   -29.006
compression  best   -31.203

ANOVA
source       SS     DF  MS      F       p       contribution
threads      31.85  2   15.92   22.895  0.0419  60.0%
batch        5.321  2   2.66    3.825   0.2073  10.0%
compression  14.53  2   7.264   10.444  0.0874  27.4%
error        1.391  2   0.6955                  2.6%
total        53.09  8

BEST SETTINGS: threads=8 batch=50 compression=none
predicted SNR at best settings: -25.327 dB
```

**RUNS.** One SNR per run, pooled over N observations. Runs 7 and 2 are the
best and worst combinations tried, 9.3 dB apart, so their RMS responses
differ by a factor of about 2.9. The grand mean is the reference for
everything below.

**MAIN EFFECTS.** Each factor's mean SNR at each level, best level starred.
Read the size of an effect as the range from its worst to its best level:
threads 4.6 dB, compression 3.0 dB, batch 1.7 dB. Threads is the knob that
matters; batch barely does.

**ANOVA.** The same ranking as numbers. Contribution is how much of the
variation between runs each factor accounts for: threads 60%, compression
27%, batch 10%, unexplained 2.6%. F and p say whether a share is more than
noise. Threads at p = 0.04 is a real effect. Compression at p = 0.09 and
batch at p = 0.21 are suggestive but not established with only two error
degrees of freedom; more observations per run or a larger array would
sharpen them. The usual thresholds are p < 0.05 for a firm effect and
p < 0.10 for a possible one. A small error row means the additive model
describes the system well. A large one means noise, or interactions between
factors, which the model does not separate. When the design is saturated
the report says so, and F and p stay unavailable until you pool.

**BEST SETTINGS and predicted SNR.** The best level of each factor and the
SNR the additive model expects there. Execute that combination as a
confirmation experiment and compare. Here it happens to be run 7, already
measured at −24.8 dB against a prediction of −25.3 dB; a gap that small,
within the error, means the model holds. A clear miss means factors
interact, and the interaction is sitting in the error row.

For a factor whose p is not small, the starred level is not a finding.
Choose that factor's level by cost or convenience.

**Notes the report can add.** If runs pooled different numbers of
observations, their SNR values are unequally precise, and the report says
so. A saturated design is flagged with the pooling advice.

### Limits

The analysis is additive: it estimates each factor on its own and leaves
interactions between factors in the error term. Noise conditions are not
laid out in an outer array; you loop over them. No confidence intervals are
computed for the predicted SNR.

## License

This is free and unencumbered software released into the public domain.
Anyone is free to copy, modify, publish, use, compile, sell, or distribute
this software, for any purpose, commercial or non-commercial, without any
conditions.
