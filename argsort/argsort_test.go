package argsort

import (
	"math/rand"
	nSort "sort"
	"testing"
)

var (
	testSample [][]float64
)

func init() {
	testSample = make([][]float64, 100)
	for i := 0; i < 100; i++ {
		size := rand.Intn(256)
		sample := make([]float64, size)
		for j := 0; j < size; j++ {
			sample[j] = float64(rand.Intn(size))
		}
		testSample[i] = sample
	}
}

func BenchmarkNativeArgSort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range testSample {
			_ = argSortedFloat(v)
		}
	}
}

func BenchmarkQuickSort(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, v := range testSample {
			_ = Quicksort(v)
		}
	}
}

func TestInsertionSort(t *testing.T) {
	var ind []int
	var c []float64

	for _, v := range testSample {
		c = make([]float64, len(v))
		ind = make([]int, len(v))
		for j := range v {
			ind[j] = j
		}
		copy(c, v)
		insertionSort(c, ind, 0, len(ind)-1)
		if !isSorted(c, ind) {
			t.Error("indices is not int sorted order", ind)
			return
		}

		for j := 0; j < len(v); j++ {
			if v[j] != c[j] {
				t.Error("sort should not change the original data", v, c)
				return
			}
		}
	}
}

func TestQuicksort(t *testing.T) {
	var ind []int
	var c []float64

	for _, v := range testSample {
		c = make([]float64, len(v))
		copy(c, v)
		ind = Quicksort(c)
		if !isSorted(c, ind) {
			t.Error("indices is not int sorted order", ind)
			return
		}

		for j := 0; j < len(v); j++ {
			if v[j] != c[j] {
				t.Error("sort should not change the original data", v, c)
			}
		}
	}
}

func isSorted(a []float64, ind []int) bool {
	for i := 1; i < len(a); i++ {
		if less(a[ind[i]], a[ind[i-1]]) {
			return false
		}
	}
	return true
}

// argSortFloat ----------
// https://gist.github.com/ericjster/1f44fda536728cbbfddd3df0e2a613d8
// argsort, like in Numpy, it returns an array of indexes into an array. Note
// that the gonum version of argsort reorders the original array and returns
// indexes to reconstruct the original order.
type argSortFloat struct {
	value   []float64 // Points to original array but does NOT alter it.
	indices []int     // Indexes to be returned.
}

func (a argSortFloat) Len() int {
	return len(a.value)
}

func (a argSortFloat) Less(i, j int) bool {
	return a.value[a.indices[i]] < a.value[a.indices[j]]
}

func (a argSortFloat) Swap(i, j int) {
	a.indices[i], a.indices[j] = a.indices[j], a.indices[i]
}

// ArgSortedFloat New allocates and returns an array of indexes into the source float array.
func argSortedFloat(src []float64) []int {
	indices := make([]int, len(src))
	for i := range src {
		indices[i] = i
	}
	a := argSortFloat{value: src, indices: indices}
	nSort.Sort(a)
	return indices
}
