package taguchi

import "errors"

// Sentinel errors. Errors returned by this package wrap one of these where
// it applies, so callers can test them with errors.Is.
var (
	// ErrNotOrthogonal reports an array in which a column, or a pair of
	// columns, is not balanced.
	ErrNotOrthogonal = errors.New("taguchi: array is not orthogonal")

	// ErrNoObservations reports a run, or an SNR input, without observations.
	ErrNoObservations = errors.New("taguchi: no observations")

	// ErrDomain reports an observation outside the domain of the chosen SNR,
	// such as a non-positive value for LargerTheBetter or a NaN.
	ErrDomain = errors.New("taguchi: observation outside the domain of the SNR")

	// ErrUndefinedSNR reports observations whose SNR is not a finite number,
	// such as a run whose observations all equal the target exactly.
	ErrUndefinedSNR = errors.New("taguchi: SNR is undefined")
)
