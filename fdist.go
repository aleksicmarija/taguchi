package taguchi

import "math"

// fPValue returns P(F > f) for an F distribution with df1 and df2 degrees
// of freedom, the probability of a variance ratio at least as large as f
// when the factor has no effect. It is NaN when f is NaN or a degree of
// freedom is not positive.
func fPValue(f float64, df1, df2 int) float64 {
	switch {
	case math.IsNaN(f) || df1 < 1 || df2 < 1:
		return math.NaN()
	case f <= 0:
		return 1
	case math.IsInf(f, 1):
		return 0
	}
	d1, d2 := float64(df1), float64(df2)
	return regularizedIncompleteBeta(d2/2, d1/2, d2/(d2+d1*f))
}

// regularizedIncompleteBeta returns I_x(a, b), evaluated by the continued
// fraction of Numerical Recipes (betacf), using the symmetry
// I_x(a, b) = 1 - I_{1-x}(b, a) where the fraction converges faster.
func regularizedIncompleteBeta(a, b, x float64) float64 {
	switch {
	case x <= 0:
		return 0
	case x >= 1:
		return 1
	}
	lgab, _ := math.Lgamma(a + b)
	lga, _ := math.Lgamma(a)
	lgb, _ := math.Lgamma(b)
	front := math.Exp(lgab - lga - lgb + a*math.Log(x) + b*math.Log(1-x))
	if x < (a+1)/(a+b+2) {
		return front * betaContinuedFraction(a, b, x) / a
	}
	return 1 - front*betaContinuedFraction(b, a, 1-x)/b
}

func betaContinuedFraction(a, b, x float64) float64 {
	const (
		maxIterations = 1000
		epsilon       = 1e-15
		tiny          = 1e-300
	)
	qab, qap, qam := a+b, a+1, a-1
	c := 1.0
	d := 1 - qab*x/qap
	if math.Abs(d) < tiny {
		d = tiny
	}
	d = 1 / d
	h := d
	for m := 1; m <= maxIterations; m++ {
		fm := float64(m)
		m2 := 2 * fm
		aa := fm * (b - fm) * x / ((qam + m2) * (a + m2))
		d = 1 + aa*d
		if math.Abs(d) < tiny {
			d = tiny
		}
		c = 1 + aa/c
		if math.Abs(c) < tiny {
			c = tiny
		}
		d = 1 / d
		h *= d * c
		aa = -(a + fm) * (qab + fm) * x / ((a + m2) * (qap + m2))
		d = 1 + aa*d
		if math.Abs(d) < tiny {
			d = tiny
		}
		c = 1 + aa/c
		if math.Abs(c) < tiny {
			c = tiny
		}
		d = 1 / d
		delta := d * c
		h *= delta
		if math.Abs(delta-1) < epsilon {
			break
		}
	}
	return h
}
