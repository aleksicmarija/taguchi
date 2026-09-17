package taguchi

import (
	"bytes"
	"math"
	"strings"
	"testing"
)

func TestReport(t *testing.T) {
	workers := NewFactor("workers", 1, 20)
	algo := NewFactor("algorithm", quick, radix)
	procs := NewFactor("procs", 4, 8)
	design, err := NewDesign(L4, workers, algo, procs)
	if err != nil {
		t.Fatal(err)
	}
	exp := NewExperiment(design)
	for i, run := range design.Runs() {
		exp.Observe(run, float64(10+3*i), float64(12+3*i))
	}
	an, err := exp.Analyze(SmallerTheBetter)
	if err != nil {
		t.Fatal(err)
	}

	first := an.String()
	for i := 0; i < 20; i++ {
		if an.String() != first {
			t.Fatal("report differs between calls")
		}
	}
	var buf bytes.Buffer
	n, err := an.WriteTo(&buf)
	if err != nil || int(n) != len(first) || buf.String() != first {
		t.Errorf("WriteTo: n=%d err=%v", n, err)
	}
	for i, line := range strings.Split(first, "\n") {
		if strings.TrimRight(line, " ") != line {
			t.Errorf("line %d has trailing spaces: %q", i, line)
		}
	}

	lines := strings.Split(first, "\n")
	if lines[0] != "Taguchi analysis: L4, 4 runs, 3 factors, smaller-the-better" {
		t.Errorf("header: %q", lines[0])
	}
	header := lines[2]
	if strings.Index(header, "workers") > strings.Index(header, "algorithm") || strings.Index(header, "algorithm") > strings.Index(header, "procs") {
		t.Errorf("factor order: %q", header)
	}
	for _, want := range []string{"\n0    1        quick", "grand mean:", "radix", "MAIN EFFECTS", "ANOVA", "n/a", "no error term", "BEST SETTINGS: workers=1 algorithm=quick procs=4", "predicted SNR"} {
		if !strings.Contains(first, want) {
			t.Errorf("report lacks %q:\n%s", want, first)
		}
	}

	pooled, err := an.Pool("procs")
	if err != nil {
		t.Fatal(err)
	}
	s := pooled.String()
	if !strings.Contains(s, "pooled into error") || strings.Contains(s, "no error term") {
		t.Errorf("pooled report:\n%s", s)
	}
}

func TestReportFormatting(t *testing.T) {
	if number(math.NaN(), "%.3f") != "n/a" || number(math.Inf(1), "%.3f") != "+inf" || number(math.Inf(-1), "%.3f") != "-inf" || number(1.23456, "%.3f") != "1.235" {
		t.Error("number")
	}
	if pValue(0.00001) != "<0.0001" || pValue(0.0421) != "0.0421" || pValue(math.NaN()) != "n/a" {
		t.Error("pValue")
	}
	if percentage(12.345) != "12.3%" || percentage(math.NaN()) != "n/a" {
		t.Error("percentage")
	}
}
