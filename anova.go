package taguchi

// computeANOVA calculates ANOVA statistics for all factors and returns:
// - ANOVAResult
// - mainEffects per factor
// - SNR per factor (same as mainEffects for convenience)
func (e *Experiment[P]) computeANOVA(oaSNR []float64, grandMean float64) (ANOVAResult, map[string][]float64, map[string][]float64) {
	oaRows := len(e.OrthogonalArray)
	totalSS := 0.0
	for _, sn := range oaSNR {
		totalSS += (sn - grandMean) * (sn - grandMean)
	}

	anova := ANOVAResult{
		FactorSS: make(map[string]float64),
		FactorDF: make(map[string]int),
		FactorMS: make(map[string]float64),
		FactorF:  make(map[string]float64),
	}

	mainEffects := map[string][]float64{}
	snrPerFactor := map[string][]float64{}

	for _, factor := range e.ControlFactors {
		levelMeans := make([]float64, len(factor.Levels))
		levelCounts := make([]int, len(factor.Levels))

		for i := 0; i < oaRows; i++ {
			levelIdx := -1
			for j, f := range e.ControlFactors {
				if f.Name == factor.Name {
					levelIdx = e.OrthogonalArray[i][j] - 1
					break
				}
			}
			if levelIdx >= 0 && levelIdx < len(factor.Levels) {
				levelMeans[levelIdx] += oaSNR[i]
				levelCounts[levelIdx]++
			}
		}

		for li := range levelMeans {
			if levelCounts[li] > 0 {
				levelMeans[li] /= float64(levelCounts[li])
			} else {
				levelMeans[li] = 0
			}
		}

		ss := 0.0
		for li := range factor.Levels {
			ss += float64(levelCounts[li]) * (levelMeans[li] - grandMean) * (levelMeans[li] - grandMean)
		}
		dfs := len(factor.Levels) - 1
		anova.FactorSS[factor.Name] = ss
		anova.FactorDF[factor.Name] = dfs
		mainEffects[factor.Name] = levelMeans
		snrPerFactor[factor.Name] = levelMeans
	}

	// Calculate error SS, DF, MS
	errorDF := oaRows - 1
	for _, df := range anova.FactorDF {
		errorDF -= df
	}
	if errorDF < 1 {
		errorDF = 1
	}

	errorSS := totalSS
	for _, ss := range anova.FactorSS {
		errorSS -= ss
	}
	errorMS := errorSS / float64(errorDF)
	anova.ErrorDF = errorDF
	anova.ErrorSS = errorSS
	anova.ErrorMS = errorMS

	// Calculate Factor MS and F-ratio
	for f, ss := range anova.FactorSS {
		df := anova.FactorDF[f]
		ms := ss / float64(df)
		anova.FactorMS[f] = ms
		anova.FactorF[f] = ms / errorMS
	}

	return anova, mainEffects, snrPerFactor
}

// computeContributions calculates the percentage contribution of each factor
// based on the ratio of its sum of squares to the total factor sum of squares.
func computeContributions(anova ANOVAResult) map[string]float64 {
	totalFactorSS := 0.0
	for _, ss := range anova.FactorSS {
		totalFactorSS += ss
	}
	contributions := map[string]float64{}
	if totalFactorSS > 0 {
		for f, ss := range anova.FactorSS {
			contributions[f] = (ss / totalFactorSS) * 100
		}
	} else {
		for f := range anova.FactorSS {
			contributions[f] = 0
		}
	}
	return contributions
}
