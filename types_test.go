package taguchi

import "testing"

func TestOptimizationGoal_String(t *testing.T) {
	tests := []struct {
		goal OptimizationGoal
		want string
	}{
		{SmallerTheBetter, "Smaller-the-Better"},
		{LargerTheBetter, "Larger-the-Better"},
		{NominalTheBest, "Nominal-the-Best"},
		{OptimizationGoal(99), "Unknown"},
	}

	for _, tt := range tests {
		got := tt.goal.String()
		if got != tt.want {
			t.Errorf("OptimizationGoal(%d).String() = %q, want %q", tt.goal, got, tt.want)
		}
	}
}
