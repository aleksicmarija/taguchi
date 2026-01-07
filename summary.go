package taguchi

import "fmt"

// PrintAnalysisReport prints a detailed, human-readable Taguchi analysis report.
func PrintAnalysisReport(result AnalysisResult) {
	fmt.Println("========================================")
	fmt.Println("        TAGUCHI ANALYSIS REPORT")
	fmt.Println("========================================")

	// 1. Optimal Factor Levels
	fmt.Println("1. Optimal Factor Levels")
	fmt.Println("------------------------")
	fmt.Println("These are the factor levels that maximize the performance metric (SNR):")
	for factor, level := range result.OptimalLevels {
		fmt.Printf("  - %s: %v\n", factor, level)
	}
	fmt.Println()

	// 2. Main Effects
	fmt.Println("2. Main Effects (Average SNR per Factor Level)")
	fmt.Println("-----------------------------------------------")
	fmt.Println("This shows how each factor level affects the response variable.")
	for factor, effects := range result.MainEffects {
		fmt.Printf("  %s:\n", factor)
		for i, val := range effects {
			fmt.Printf("    Level %d: %.4f\n", i+1, val)
		}
		fmt.Println("    => Higher values indicate a better effect on performance.")
	}

	// 3. Contributions of Each Factor
	fmt.Println("3. Contribution of Each Factor")
	fmt.Println("-------------------------------")
	fmt.Println("This tells us how much each factor contributes to the total variation:")
	for factor, contrib := range result.Contributions {
		fmt.Printf("  - %s: %.2f%%\n", factor, contrib)
	}
	fmt.Println("  => Factors with higher percentages are more influential.")

	// 4. ANOVA Results
	fmt.Println("4. ANOVA (Analysis of Variance) Table")
	fmt.Println("------------------------------------")
	fmt.Println("ANOVA helps determine which factors significantly affect the response.")
	fmt.Printf("%-15s %-12s %-8s %-10s\n", "Factor", "SS", "DF", "F-ratio")
	for factor := range result.ANOVA.FactorSS {
		fmt.Printf("%-15s %-12.4f %-8d %-10.4f\n",
			factor,
			result.ANOVA.FactorSS[factor],
			result.ANOVA.FactorDF[factor],
			result.ANOVA.FactorF[factor],
		)
	}
	fmt.Printf("%-15s %-12.4f %-8d\n",
		"Error",
		result.ANOVA.ErrorSS,
		result.ANOVA.ErrorDF,
	)
	fmt.Println("  => Factors with higher F-ratio are more statistically significant.")
}
