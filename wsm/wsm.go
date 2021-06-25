package main

import (
	"color-thief/argsort"
	"color-thief/helper"
	"color-thief/rgbUtil"
	"color-thief/wu"
	"fmt"
	"log"
	"math/rand"
	"time"
)

const (
	HistBits = 5
	Shift    = 8 - HistBits
	HistSize = 1 << (3 * HistBits)
)

func getHistogram(pixels [][]int) ([]float64, [][]float64) {
	pix := helper.New2dMatrixFloat(HistSize, 3)
	hist := make([]float64, HistSize)

	var index, r, g, b int
	for _, p := range pixels {
		r = p[0] >> Shift
		g = p[1] >> Shift
		b = p[2] >> Shift
		index = (r << (2 * HistBits)) + (g << HistBits) + b
		pix[index] = convertToFloat64(p)
		hist[index]++
	}
	// normalize weight by the number of pixels in the image
	for i := 0; i < HistSize; i++ {
		hist[i] /= float64(len(pixels))
	}
	return hist, pix
}

// Todo fix
func WSM(src [][]int, k int) {
	// variables
	var centroids [][]float64 // centroid list with size of k
	var groups [][]float64    // use when computing new centroids
	var clusterSize []float64 // cluster size
	var d [][]float64         // distance matrix
	var m [][]int             // distance rank matrix
	var p2c []int             // pointer to centroid index
	var p int
	var dist, minDist, prevDist, wcss float64

	var hist []float64
	var pixels [][]float64

	// get histogram
	hist, pixels = getHistogram(src)

	// select init cluster centers
	centroids = helper.New2dMatrixFloat(k, 3)
	for i, v := range wu.QuantWu(src, k) {
		centroids[i] = convertToFloat64(v)
	}

	// random assign centroid for each pixels
	p2c = make([]int, len(hist))
	for i := range hist {
		p2c[i] = rand.Intn(k)
	}

	// default 100 iterations for k-means
	for iter := 0; iter < 100; iter++ {
		// compute distance matrix
		d = helper.New2dMatrixFloat(k, k)
		for i := 0; i < k; i++ {
			for j := i + 1; j < k; j++ {
				dist = distance(centroids[i], centroids[j])
				d[i][j], d[j][i] = dist, dist
			}
		}

		// compute distance rank matrix
		m = rank(d, k)

		for i, w := range hist {
			if w == 0 {
				continue
			}
			p = p2c[i]
			dist = distance(pixels[i], centroids[p])
			minDist = dist
			prevDist = dist
			for j := 1; j < k; j++ {
				t := m[p][j]
				if d[p][t] >= 4*prevDist {
					break
				}
				dist = distance(pixels[i], centroids[t])
				if dist < minDist {
					minDist = dist
					p2c[i] = t
				}
			}
		}

		groups = helper.New2dMatrixFloat(k, 4)
		clusterSize = make([]float64, k)
		// recalculate the cluster centers
		for i, w := range hist {
			if w == 0 {
				continue
			}
			p = p2c[i]
			groups[p][0] += pixels[i][0] * w // r
			groups[p][1] += pixels[i][1] * w // g
			groups[p][2] += pixels[i][2] * w // b
			groups[p][3] += w
			clusterSize[p]++
		}

		wcss = 0
		for i, c := range groups {
			nR := c[0] / c[3]
			nG := c[1] / c[3]
			nB := c[2] / c[3]

			wcss += (centroids[i][0]-nR)*(centroids[i][0]-nR) +
				(centroids[i][1]-nG)*(centroids[i][1]-nG) +
				(centroids[i][2]-nB)*(centroids[i][2]-nB)
			centroids[i] = []float64{nR, nG, nB}
		}

		fmt.Println(wcss)
		if wcss < 1e-3 {
			fmt.Println("iterations:", iter)
			break
		}
	}
	//
	//clusterRank := argsort.ArgSortedFloat(clusterSize)
	//sb := strings.Builder{}
	//for i := k - 1; i >= 0; i-- {
	//	c := centroids[clusterRank[i]]
	//	sb.WriteString(fmt.Sprintf("\"#%02x%02x%02x\"", uint8(c[0]), uint8(c[1]), uint8(c[2])))
	//	if i > 0 {
	//		sb.WriteString(",")
	//	}
	//}
	//fmt.Println(sb.String())
}

// Construct a K × K matrix M in which row i is a permutation of 1, 2, . . . , K that
// represents the clusters in increasing order of distance of their centers from c_i
func rank(d [][]float64, k int) [][]int {
	m := helper.New2dMatrixInt(k, k)
	for i, v := range d {
		m[i] = argsort.ArgSortedFloat(v)
	}
	return m
}

func distance(c1, c2 []float64) float64 {
	sum := 0.0
	for i := range c1 {
		sum += (c1[i] - c2[i]) * (c1[i] - c2[i])
	}
	return sum
}

func convertToFloat64(p []int) []float64 {
	c := make([]float64, 3)
	for i, v := range p {
		c[i] = float64(v)
	}
	return c
}

func main() {
	img1, err := rgbUtil.ReadImage("example/photo1.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	start := time.Now()
	p := helper.SubsamplingPixels(img1)
	fmt.Println(time.Since(start))
	start = time.Now()
	WSM(p, 6)
	fmt.Println(time.Since(start))
}
