package taguchi

import "testing"

func TestStandardArrays_Exist(t *testing.T) {
	expected := []ArrayType{L4, L8, L9, L16, L18}
	for _, name := range expected {
		if _, ok := StandardArrays[name]; !ok {
			t.Errorf("StandardArrays missing %q", name)
		}
	}
}

func TestStandardArrays_Dimensions(t *testing.T) {
	expectedRows := map[ArrayType]int{
		L4:  4,
		L8:  8,
		L9:  9,
		L16: 16,
		L18: 18,
	}

	for name, wantRows := range expectedRows {
		arr := StandardArrays[name]
		if len(arr) != wantRows {
			t.Errorf("StandardArrays[%q] has %d rows, want %d", name, len(arr), wantRows)
		}
	}
}

func TestStandardArrays_ColumnConsistency(t *testing.T) {
	for name, arr := range StandardArrays {
		if len(arr) == 0 {
			t.Errorf("StandardArrays[%q] is empty", name)
			continue
		}
		expectedCols := len(arr[0])
		for i, row := range arr {
			if len(row) != expectedCols {
				t.Errorf("StandardArrays[%q] row %d has %d columns, want %d", name, i, len(row), expectedCols)
			}
		}
	}
}

func TestStandardArrays_ValueBounds(t *testing.T) {
	for name, arr := range StandardArrays {
		for i, row := range arr {
			for j, val := range row {
				if val < 1 {
					t.Errorf("StandardArrays[%q][%d][%d] = %d, want >= 1", name, i, j, val)
				}
			}
		}
	}
}

func TestStandardArrays_Balance(t *testing.T) {
	for name, arr := range StandardArrays {
		rows := len(arr)
		if rows == 0 {
			continue
		}
		cols := len(arr[0])

		for col := 0; col < cols; col++ {
			// Count occurrences of each level in this column
			counts := make(map[int]int)
			for row := 0; row < rows; row++ {
				counts[arr[row][col]]++
			}

			// All levels should appear the same number of times
			var expectedCount int
			first := true
			for _, count := range counts {
				if first {
					expectedCount = count
					first = false
				} else if count != expectedCount {
					t.Errorf("StandardArrays[%q] column %d is not balanced: level counts = %v",
						name, col, counts)
					break
				}
			}
		}
	}
}
