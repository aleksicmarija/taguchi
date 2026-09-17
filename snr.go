package taguchi

import (
	"fmt"
	"math"
)

// SNR is a signal-to-noise ratio: a named function from the pooled
// observations of one run to a value in decibels. Higher is better for
// every SNR in this package. Name appears in reports.
//
// Of must refuse input for which the ratio is not defined rather than
// substitute a value: the analysis of variance cannot recover from an
// infinite or fabricated number. The ratios in this package return an error
// wrapping ErrNoObservations, ErrDomain or ErrUndefinedSNR.
//
// A custom quality characteristic is an SNR literal:
//
//	p99 := taguchi.SNR{Name: "p99 latency", Of: func(y []float64) (float64, error) { ... }}
type SNR struct {
	Name string
	Of   func(y []float64) (float64, error)
}

// SmallerTheBetter is the SNR for a response that should be as small as
// possible, such as latency or defect count:
//
//	SNR = -10 log10( mean(y²) )
//
// Observations may be any finite number. The SNR is undefined when every
// observation is zero.
var SmallerTheBetter = SNR{Name: "smaller-the-better", Of: smallerTheBetter}

// LargerTheBetter is the SNR for a response that should be as large as
// possible, such as throughput or yield:
//
//	SNR = -10 log10( mean(1/y²) )
//
// Every observation must be positive.
var LargerTheBetter = SNR{Name: "larger-the-better", Of: largerTheBetter}

// NominalTheBest is Taguchi's SNR for a response that should sit on a target
// with as little variation as possible:
//
//	SNR = 10 log10( (Sm - Ve) / (n Ve) )    with Sm = n mean(y)², Ve = var(y)
//	    = 10 log10( mean(y)² / var(y) - 1/n )
//
// var is the sample variance with n-1 in the denominator. Subtracting
// var(y)/n makes the numerator an unbiased estimate of the squared mean.
// Minitab and JMP omit that term; the two differ by less than 0.05 dB once
// the mean is more than ten standard errors from zero.
//
// The ratio does not depend on the target, which supports two-step
// optimisation: first choose the levels that maximise the SNR, then move
// the mean onto the target with a factor that affects only the mean. Use
// NominalTheBestTarget to penalise distance from the target instead.
//
// A run needs at least two observations, a non-zero variance, and a mean
// more than one standard error away from zero. A response centred near
// zero should use NominalTheBestTarget.
var NominalTheBest = SNR{Name: "nominal-the-best", Of: nominalTheBest}

// NominalTheBestTarget returns the SNR for a response that should be as
// close as possible to target, penalising bias and variation together:
//
//	SNR = -10 log10( mean((y - target)²) )
//
// The SNR is undefined when every observation equals the target exactly.
func NominalTheBestTarget(target float64) SNR {
	return SNR{
		Name: fmt.Sprintf("nominal-the-best, target %v", target),
		Of: func(y []float64) (float64, error) {
			if err := checkObservations(y); err != nil {
				return math.NaN(), err
			}
			if math.IsNaN(target) || math.IsInf(target, 0) {
				return math.NaN(), fmt.Errorf("%w: target is %v", ErrDomain, target)
			}
			sum := 0.0
			for _, v := range y {
				d := v - target
				sum += d * d
			}
			return decibels(sum / float64(len(y)))
		},
	}
}

func smallerTheBetter(y []float64) (float64, error) {
	if err := checkObservations(y); err != nil {
		return math.NaN(), err
	}
	sum := 0.0
	for _, v := range y {
		sum += v * v
	}
	return decibels(sum / float64(len(y)))
}

func largerTheBetter(y []float64) (float64, error) {
	if err := checkObservations(y); err != nil {
		return math.NaN(), err
	}
	sum := 0.0
	for i, v := range y {
		if v <= 0 {
			return math.NaN(), fmt.Errorf("%w: larger-the-better needs positive observations, y[%d] = %v", ErrDomain, i, v)
		}
		sum += 1 / (v * v)
	}
	return decibels(sum / float64(len(y)))
}

func nominalTheBest(y []float64) (float64, error) {
	if err := checkObservations(y); err != nil {
		return math.NaN(), err
	}
	n := float64(len(y))
	if n < 2 {
		return math.NaN(), fmt.Errorf("%w: nominal-the-best needs at least 2 observations to estimate variance, got %d", ErrUndefinedSNR, len(y))
	}
	mean := 0.0
	for _, v := range y {
		mean += v
	}
	mean /= n
	ss := 0.0
	for _, v := range y {
		ss += (v - mean) * (v - mean)
	}
	variance := ss / (n - 1)
	if variance == 0 {
		return math.NaN(), fmt.Errorf("%w: nominal-the-best with zero variance", ErrUndefinedSNR)
	}
	signal := mean*mean - variance/n
	if signal <= 0 {
		return math.NaN(), fmt.Errorf("%w: nominal-the-best: mean %v is within one standard error of zero; use NominalTheBestTarget", ErrUndefinedSNR, mean)
	}
	return finite(10 * math.Log10(signal/variance))
}

func checkObservations(y []float64) error {
	if len(y) == 0 {
		return ErrNoObservations
	}
	for i, v := range y {
		if math.IsNaN(v) || math.IsInf(v, 0) {
			return fmt.Errorf("%w: y[%d] = %v", ErrDomain, i, v)
		}
	}
	return nil
}

// decibels converts a mean squared deviation to -10 log10(msd).
func decibels(msd float64) (float64, error) {
	if msd == 0 {
		return math.NaN(), fmt.Errorf("%w: mean squared deviation is zero", ErrUndefinedSNR)
	}
	return finite(-10 * math.Log10(msd))
}

func finite(snr float64) (float64, error) {
	if math.IsNaN(snr) || math.IsInf(snr, 0) {
		return math.NaN(), fmt.Errorf("%w: result is %v", ErrUndefinedSNR, snr)
	}
	return snr, nil
}
