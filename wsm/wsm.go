package wsm

import (
	"color-thief/argsort"
	"color-thief/wu"
	"math"
)

const (
	HistBits = 5
	Shift    = 8 - HistBits
	HistSize = 1 << (3 * HistBits)
)

func getHistogram(pixels [][3]int) ([HistSize]float64, [HistSize][3]float64) {
	var ind, r, g, b, i int
	var inr, ing, inb int
	var size float64

	pix := [HistSize][3]float64{}
	hist := [HistSize]float64{}

	for i = range pixels {
		r = pixels[i][0]
		g = pixels[i][1]
		b = pixels[i][2]

		inr = r >> Shift
		ing = g >> Shift
		inb = b >> Shift

		ind = (inr << (2 * HistBits)) + (ing << HistBits) + inb
		pix[ind][0], pix[ind][1], pix[ind][2] = float64(r), float64(g), float64(b)
		hist[ind]++
	}

	// normalize weight by the number of pixels in the image
	size = float64(len(pixels))
	for i = 0; i < HistSize; i++ {
		hist[i] /= size
	}
	return hist, pix
}

func WSM(src [][3]int, k int) ([][3]int, error) {
	// variables
	var centroids [][3]float64          // centroid list with size of k
	var d []float64                     // distance matrix
	var m []int                         // distance rank matrix
	var hist [HistSize]float64          // image encoded histogram
	var pixels [HistSize][3]float64     // encoded unique pixels
	var p2c [HistSize]int               // pointer to centroid index
	var cR, cG, cB, cW, cSize []float64 // use when computing new centroids
	var nR, nG, nB float64              // new centroid r,g,b
	var palette [][3]int                // palette container
	var cPix [3]float64                 // pixel with float
	var pix [3]int                      // pixel with int
	var dist, minDist, prevDist float64
	var loss, tempLoss float64
	var size, w float64
	var iter, i, j int
	var p, t int
	var err error

	// get histogram
	hist, pixels = getHistogram(src)

	// init cluster centers based on wu color quantization result
	palette, err = wu.QuantWu(src, k)
	if err != nil {
		return nil, err
	}
	// cannot produce enough color, create palette using color scheme
	if len(palette) < k {
		return palette, nil
	}

	// init centroids
	centroids = make([][3]float64, k)
	for i, pix = range palette {
		centroids[i][0], centroids[i][1], centroids[i][2] = float64(pix[0]), float64(pix[1]), float64(pix[2])
	}

	// random assign centroids to each pixels
	for i = 0; i < HistSize; i++ {
		if hist[i] == 0 {
			continue
		}
		p2c[i] = i % k
	}

	loss = 1e6
	size = float64(len(src))
	d = make([]float64, k*k)
	m = make([]int, k*k)
	cR = make([]float64, k)
	cG = make([]float64, k)
	cB = make([]float64, k)
	cW = make([]float64, k)
	cSize = make([]float64, k)
	// default 100 iterations for k-means
	for iter = 0; iter < 100; iter++ {
		// compute distance matrix
		for i = 0; i < k; i++ {
			for j = i + 1; j < k; j++ {
				dist = (centroids[i][0]-centroids[j][0])*(centroids[i][0]-centroids[j][0]) +
					(centroids[i][1]-centroids[j][1])*(centroids[i][1]-centroids[j][1]) +
					(centroids[i][2]-centroids[j][2])*(centroids[i][2]-centroids[j][2])
				dist = math.Sqrt(dist)
				d[i*k+j], d[j*k+i] = dist, dist
			}
		}

		// Construct a K × K matrix M in which row i is a permutation of 1, 2, . . . , K that
		// represents the clusters in increasing order of distance of their centers from c_i
		for i = 0; i < k; i++ {
			r, err := argsort.ArgSortedFloat(d[i*k : i*k+k])
			if err != nil {
				return nil, err
			}
			copy(m[i*k:i*k+k], r)
		}

		for i, w = range hist {
			if w == 0 {
				continue
			}
			p = p2c[i]
			cPix = pixels[i]
			dist = (cPix[0]-centroids[p][0])*(cPix[0]-centroids[p][0]) +
				(cPix[1]-centroids[p][1])*(cPix[1]-centroids[p][1]) +
				(cPix[2]-centroids[p][2])*(cPix[2]-centroids[p][2])
			dist = math.Sqrt(dist)
			minDist, prevDist = dist, dist
			for j = 1; j < k; j++ {
				t = m[p*k+j]
				if d[p*k+t] >= 4*prevDist {
					break // There can be no other closer center. Stop checking
				}
				dist = (cPix[0]-centroids[t][0])*(cPix[0]-centroids[t][0]) +
					(cPix[1]-centroids[t][1])*(cPix[1]-centroids[t][1]) +
					(cPix[2]-centroids[t][2])*(cPix[2]-centroids[t][2])
				dist = math.Sqrt(dist)
				if dist <= minDist {
					minDist = dist
					p2c[i] = t
				}
			}
		}

		// reset matrix
		for i = 0; i < k; i++ {
			cR[i], cG[i], cB[i], cW[i], cSize[i] = 0, 0, 0, 0, 0
		}

		// recalculate the cluster centers
		for i, w = range hist {
			if w == 0 {
				continue
			}
			p = p2c[i]
			cR[p] += pixels[i][0] * w // r
			cG[p] += pixels[i][1] * w // g
			cB[p] += pixels[i][2] * w // b
			cW[p] += w
			cSize[p] += w * size
		}

		for i = 0; i < k; i++ {
			nR = cR[i] / cW[i]
			nG = cG[i] / cW[i]
			nB = cB[i] / cW[i]

			centroids[i][0], centroids[i][1], centroids[i][2] = nR, nG, nB
		}

		tempLoss = 0
		for i, w = range hist {
			if w == 0 {
				continue
			}
			p = p2c[i]
			cPix = pixels[i]
			dist = (cPix[0]-centroids[p][0])*(cPix[0]-centroids[p][0]) +
				(cPix[1]-centroids[p][1])*(cPix[1]-centroids[p][1]) +
				(cPix[2]-centroids[p][2])*(cPix[2]-centroids[p][2])
			dist = math.Sqrt(dist)
			tempLoss += dist
		}

		if loss-tempLoss < 1e-3 {
			break
		}
		loss = tempLoss
	}

	clusterRank, err := argsort.ArgSortedFloat(cSize)
	if err != nil {
		return nil, err
	}
	for i = 0; i < k; i++ {
		cPix = centroids[clusterRank[k-1-i]]
		palette[i][0], palette[i][1], palette[i][2] = int(cPix[0]), int(cPix[1]), int(cPix[2])
	}
	return palette, nil
}
