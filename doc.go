// Package taguchi designs and analyses Taguchi (robust design) experiments.
//
// An experiment has three stages.
//
// Design. Declare each control factor with NewFactor, giving it a name and
// its levels, and assign the factors to the columns of an orthogonal array
// with NewDesign. Level values may be of any type: an int, a duration, an
// enum, a struct. The library only ever works with level indices.
//
//	workers := taguchi.NewFactor("workers", 1, 4, 16)
//	algo    := taguchi.NewFactor("algorithm", QuickSort, RadixSort, MergeSort)
//	design, err := taguchi.NewDesign(taguchi.L9, workers, algo)
//
// Run. Execute every Run of the design under whatever noise conditions you
// choose and record each measured response with Experiment.Observe. A typed
// factor returns its level for a run with Factor.Of.
//
//	exp := taguchi.NewExperiment(design)
//	for _, run := range design.Runs() {
//		for _, input := range inputs { // noise conditions
//			exp.Observe(run, measure(algo.Of(run), workers.Of(run), input))
//		}
//	}
//
// Analyse. Choose the signal-to-noise ratio that matches the quality
// characteristic and call Experiment.Analyze. All observations of a run are
// pooled into one SNR value, and the SNR values are decomposed into main
// effects and an analysis of variance.
//
//	analysis, err := exp.Analyze(taguchi.SmallerTheBetter)
//	fmt.Print(analysis)
//	best := algo.Best(analysis)
//
// Noise factors are deliberately not modelled: the analysis pools every
// observation of a run, so the library does not need to know how the
// observations were produced. Enumerate noise conditions in your own loop.
//
// Every index in the API is zero-based: runs, columns, levels and the Best
// level of an Effect can all be used directly as slice indices. The one
// exception is the input of NewOrthogonalArray, which accepts tables in the
// form printed in reference books, with levels numbered from 1.
package taguchi
