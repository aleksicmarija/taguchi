package taguchi

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"
)

// WriteTo writes a plain-text report: the SNR of every run, the main
// effects, the ANOVA table and the best settings. Factors appear in design
// order, so the output is the same on every call.
func (a *Analysis) WriteTo(w io.Writer) (int64, error) {
	var buf bytes.Buffer
	a.report(&buf)
	n, err := w.Write(buf.Bytes())
	return int64(n), err
}

// String returns the report that WriteTo writes.
func (a *Analysis) String() string {
	var b strings.Builder
	a.report(&b)
	return b.String()
}

func (a *Analysis) report(w io.Writer) {
	fmt.Fprintf(w, "Taguchi analysis: %s, %d runs, %d factors", a.arrayName(), len(a.Runs), len(a.Effects))
	if a.SNR.Name != "" {
		fmt.Fprintf(w, ", %s", a.SNR.Name)
	}
	fmt.Fprint(w, "\n\n")

	fmt.Fprintln(w, "RUNS (SNR in dB, pooled over N observations)")
	header := []string{"run"}
	for _, e := range a.Effects {
		header = append(header, e.Factor)
	}
	rows := [][]string{append(header, "N", "SNR")}
	for _, r := range a.Runs {
		row := []string{strconv.Itoa(r.Row)}
		for _, e := range a.Effects {
			row = append(row, e.Levels[a.array.Level(r.Row, e.column)])
		}
		n := "-"
		if r.N > 0 {
			n = strconv.Itoa(r.N)
		}
		rows = append(rows, append(row, n, number(r.SNR, "%.3f")))
	}
	writeTable(w, rows)
	fmt.Fprintf(w, "grand mean: %s dB\n", number(a.Mean, "%.3f"))
	if !a.Balanced() {
		lo, hi := a.Runs[0].N, a.Runs[0].N
		for _, r := range a.Runs {
			lo, hi = min(lo, r.N), max(hi, r.N)
		}
		fmt.Fprintf(w, "note: runs pooled between %d and %d observations, so their SNR values are not\n", lo, hi)
		fmt.Fprintln(w, "      equally precise. Observe every run the same number of times if you can.")
	}
	fmt.Fprintln(w)

	fmt.Fprintln(w, "MAIN EFFECTS (mean SNR per level, dB; * marks the best level)")
	rows = [][]string{{"factor", "level", "mean SNR"}}
	for _, e := range a.Effects {
		for l, m := range e.Means {
			mark := ""
			if l == e.Best {
				mark = "*"
			}
			rows = append(rows, []string{e.Factor, e.Levels[l], number(m, "%.3f"), mark})
		}
	}
	writeTable(w, rows)
	fmt.Fprintln(w)

	fmt.Fprintln(w, "ANOVA")
	rows = [][]string{{"source", "SS", "DF", "MS", "F", "p", "contribution"}}
	for _, e := range a.Effects {
		note := ""
		if e.Pooled {
			note = "pooled into error"
		}
		rows = append(rows, []string{
			e.Factor, number(e.SS, "%.4g"), strconv.Itoa(e.DF), number(e.MS, "%.4g"),
			number(e.F, "%.3f"), pValue(e.P), percentage(e.Contribution), note,
		})
	}
	rows = append(rows,
		[]string{"error", number(a.Error.SS, "%.4g"), strconv.Itoa(a.Error.DF), number(a.Error.MS, "%.4g"), "", "", percentage(a.Error.Contribution)},
		[]string{"total", number(a.TotalSS, "%.4g"), strconv.Itoa(a.TotalDF)},
	)
	writeTable(w, rows)
	if a.Saturated() {
		fmt.Fprintln(w, "note: the factors use every degree of freedom, so there is no error term and F and p")
		fmt.Fprintln(w, "      are not available. Pool the least significant factors with Analysis.Pool.")
	}
	fmt.Fprintln(w)

	fmt.Fprint(w, "BEST SETTINGS:")
	for _, e := range a.Effects {
		fmt.Fprintf(w, " %s=%s", e.Factor, e.Levels[e.Best])
	}
	fmt.Fprintf(w, "\npredicted SNR at best settings: %s dB\n", number(a.PredictedSNR(), "%.3f"))
}

// arrayName names the design's array when there is one, since the analysis
// itself only holds the columns that carry factors.
func (a *Analysis) arrayName() string {
	if a.design != nil {
		return a.design.array.Name()
	}
	return a.array.Name()
}

// writeTable writes rows as aligned columns without trailing spaces.
func writeTable(w io.Writer, rows [][]string) {
	var buf bytes.Buffer
	tw := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t")+"\t")
	}
	tw.Flush()
	for _, line := range strings.SplitAfter(buf.String(), "\n") {
		if line == "" {
			continue
		}
		fmt.Fprintln(w, strings.TrimRight(line, " \n"))
	}
}

func number(x float64, format string) string {
	switch {
	case math.IsNaN(x):
		return "n/a"
	case math.IsInf(x, 1):
		return "+inf"
	case math.IsInf(x, -1):
		return "-inf"
	}
	return fmt.Sprintf(format, x)
}

func pValue(p float64) string {
	if !math.IsNaN(p) && p < 0.0001 {
		return "<0.0001"
	}
	return number(p, "%.4f")
}

func percentage(p float64) string {
	if math.IsNaN(p) {
		return "n/a"
	}
	return fmt.Sprintf("%.1f%%", p)
}
