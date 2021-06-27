package argsort

import (
	"testing"
)

var (
	data      [][]float64
	unchanged [][]float64
	expected  [][]int
)

func init() {
	data = [][]float64{
		{1, 25, 3, 5, 4},
		{1, 5, 4, 25, 3},
		{1, 25, 5, 4, 3},
		{1, 4, 3, 25, 5},
		{3, 5, 1, 25, 4},
	}

	unchanged = [][]float64{
		{1, 25, 3, 5, 4},
		{1, 5, 4, 25, 3},
		{1, 25, 5, 4, 3},
		{1, 4, 3, 25, 5},
		{3, 5, 1, 25, 4},
	}

	expected = [][]int{
		{0, 2, 4, 3, 1},
		{0, 4, 2, 1, 3},
		{0, 4, 3, 2, 1},
		{0, 2, 1, 4, 3},
		{2, 0, 4, 1, 3},
	}
}

func TestArgSort(t *testing.T) {
	for i := 0; i < len(data); i++ {
		res, err := ArgSortedFloat(data[i])
		if err != nil {
			t.Error(err)
		}
		for j := 0; j < len(data[i]); j++ {
			if expected[i][j] != res[j] {
				t.Error("unequaled result", expected[i], res)
			}
		}
	}

	for i := 0; i < len(data); i++ {
		for j := 0; j < len(data[i]); j++ {
			if unchanged[i][j] != data[i][j] {
				t.Error("sort should not change the original data", unchanged[i], data[i])
			}
		}
	}
}

func BenchmarkArgSort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range data {
			_, _ = ArgSortedFloat(v)
		}
	}
}
