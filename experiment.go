package taguchi

import (
	"fmt"
	"slices"
	"sync"
)

// Experiment collects the observations made while running a Design. It is
// safe for concurrent use, so runs may be executed in parallel.
type Experiment struct {
	mu     sync.Mutex
	design *Design
	obs    [][]float64
}

// NewExperiment returns an empty experiment for design.
func NewExperiment(design *Design) *Experiment {
	if design == nil {
		panic("taguchi: NewExperiment: nil design")
	}
	return &Experiment{design: design, obs: make([][]float64, design.array.Runs())}
}

// Design returns the design the experiment runs.
func (e *Experiment) Design() *Design { return e.design }

// Observe records responses measured in run. A run may be observed any
// number of times, typically once per noise condition and repetition; the
// analysis pools all of them. It panics if run belongs to another design.
func (e *Experiment) Observe(run Run, y ...float64) {
	e.check(run)
	e.mu.Lock()
	defer e.mu.Unlock()
	e.obs[run.row] = append(e.obs[run.row], y...)
}

// Observations returns a copy of the responses recorded for run.
func (e *Experiment) Observations(run Run) []float64 {
	e.check(run)
	e.mu.Lock()
	defer e.mu.Unlock()
	return slices.Clone(e.obs[run.row])
}

func (e *Experiment) check(run Run) {
	if run.design != e.design {
		panic(fmt.Sprintf("taguchi: %v does not belong to this experiment's design", run))
	}
}

// Analyze pools the observations of every run into one value with snr and
// decomposes those values over the factors of the design.
//
// It returns an error wrapping ErrNoObservations if any run has no
// observations, and passes through the error of snr for any run it rejects.
// A design whose factors use every degree of freedom has no error term; the
// analysis still succeeds, with Saturated reporting true, and Pool can be
// used to obtain one.
func (e *Experiment) Analyze(snr SNR) (*Analysis, error) {
	if snr.Of == nil {
		return nil, fmt.Errorf("taguchi: Analyze: SNR %q has no function", snr.Name)
	}
	e.mu.Lock()
	obs := make([][]float64, len(e.obs))
	for i, ys := range e.obs {
		obs[i] = slices.Clone(ys)
	}
	e.mu.Unlock()

	var missing []int
	for i, ys := range obs {
		if len(ys) == 0 {
			missing = append(missing, i)
		}
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("%w: runs %v have no observations", ErrNoObservations, missing)
	}

	values := make([]float64, len(obs))
	for i, ys := range obs {
		v, err := snr.Of(ys)
		if err != nil {
			return nil, fmt.Errorf("run %d: %w", i, err)
		}
		values[i] = v
	}

	a, err := AnalyzeSNR(e.design.factorArray(), values)
	if err != nil {
		return nil, err
	}
	a.design = e.design
	a.SNR = snr
	for i := range a.Runs {
		a.Runs[i].N = len(obs[i])
	}
	for j, f := range e.design.factors {
		eff := &a.Effects[j]
		eff.Factor = f.Name()
		for l := range eff.Levels {
			eff.Levels[l] = f.label(l)
		}
	}
	return a, nil
}
