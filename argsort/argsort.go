package argsort

import (
	"errors"
	"sort"
)

// ArgSortedFloat ----------
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
func ArgSortedFloat(src []float64) ([]int, error) {
	indices := make([]int, len(src))
	for i := range src {
		indices[i] = i
	}
	if len(src) != len(indices) {
		return nil, errors.New("floats: length of indices does not match length of slice")
	}
	a := argSortFloat{value: src, indices: indices}
	sort.Sort(a)
	return indices, nil
}

type argSortInt struct {
	value   []int // Points to original array but does NOT alter it.
	indices []int // Indexes to be returned.
}

func (a argSortInt) Len() int {
	return len(a.value)
}

func (a argSortInt) Less(i, j int) bool {
	return a.value[a.indices[i]] < a.value[a.indices[j]]
}

func (a argSortInt) Swap(i, j int) {
	a.indices[i], a.indices[j] = a.indices[j], a.indices[i]
}

// ArgSortedInt New allocates and returns an array of indexes into the source float array.
func ArgSortedInt(src []int) ([]int, error) {
	indices := make([]int, len(src))
	for i := range src {
		indices[i] = i
	}
	if len(src) != len(indices) {
		return nil, errors.New("floats: length of indices does not match length of slice")
	}
	a := argSortInt{value: src, indices: indices}
	sort.Sort(a)
	return indices, nil
}
