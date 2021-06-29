package argsort

/**
Implement argsort using optimized quicksort and insertion sort mentioned in
Algorithms, 4th Edition by Robert Sedgewick and Kevin Wayne
*/

const insertionSortCutoff = 8

func Quicksort(a []float64) []int {
	var ind []int

	ind = make([]int, len(a))
	for i := range a {
		ind[i] = i
	}
	sort(a, ind, 0, len(a)-1)
	return ind
}

// quicksort the subarray from a[lo] to a[hi]
func sort(a []float64, ind []int, lo, hi int) {
	var n, j int
	if hi <= lo {
		return
	}

	// cutoff to insertion sort (Insertion.sort() uses half-open intervals)
	n = hi - lo + 1
	if n <= insertionSortCutoff {
		insertionSort(a, ind, lo, hi)
		return
	}

	j = partition(a, ind, lo, hi)
	sort(a, ind, lo, j-1)
	sort(a, ind, j+1, hi)
}

// insertionSort optimized with half exchanges and a sentinel
func insertionSort(a []float64, ind []int, lo, hi int) {
	var i, j, k, exchanges int

	// put smallest element in position to serve as sentinel
	exchanges = 0
	for i = hi; i > lo; i-- {
		if less(a[ind[i]], a[ind[i-1]]) {
			swap(ind, i, i-1)
			exchanges++
		}
	}
	if exchanges == 0 {
		return
	}

	// insertion sort with half-exchanges
	for i = lo + 1; i <= hi; i++ {
		k = ind[i]
		j = i
		for less(a[k], a[ind[j-1]]) {
			ind[j] = ind[j-1]
			j--
		}
		ind[j] = k
	}
}

func partition(a []float64, ind []int, lo, hi int) int {
	var n, m, i, j, k int

	n = hi - lo + 1
	m = median3(a, ind, lo, lo+n/2, hi)
	swap(ind, m, lo)

	i = lo + 1
	j = hi

	k = ind[lo]

	// a[lo] is unique largest element
	for less(a[ind[i]], a[k]) {
		if i == hi {
			swap(ind, lo, hi)
			return hi
		}
		i++
	}

	// a[lo] is unique smallest element
	for less(a[k], a[ind[j]]) {
		if j == lo+1 {
			return lo
		}
		j--
	}

	// the main loop
	for i < j {
		swap(ind, i, j)
		i++
		for less(a[ind[i]], a[k]) {
			i++
		}
		j--
		for less(a[k], a[ind[j]]) {
			j--
		}
	}

	// put partitioning item v at a[j]
	swap(ind, lo, j)

	// now, a[lo .. j-1] <= a[j] <= a[j+1 .. hi]
	return j
}

func median3(a []float64, ind []int, i, j, k int) int {
	indI, indJ, indK := ind[i], ind[j], ind[k]
	if less(a[indI], a[indJ]) {
		if less(a[indJ], a[indK]) {
			return j
		}
		if less(a[indI], a[indK]) {
			return k
		}
		return i
	}
	if less(a[indK], a[indJ]) {
		return j
	}
	if less(a[indK], a[indI]) {
		return k
	}
	return i
}

// is v < w ?
func less(v, w float64) bool {
	return v < w
}

func swap(a []int, i, j int) {
	a[i], a[j] = a[j], a[i]
}
