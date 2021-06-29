package wu

import (
	"color-thief/argsort"
)

/**********************************************************************
		Go Implementation of Wu's Color Quantizer (v. 2)
			(see Graphics Gems vol. II, pp. 126-133)
	Ported and modified from: https://gist.github.com/bert/1192520
**********************************************************************/

const (
	maxColor = 256
	red      = 2
	green    = 1
	blue     = 0
	cubeSize = 33 * 33 * 33
)

type box struct {
	r0  int /* min value, exclusive */
	r1  int /* max value, inclusive */
	g0  int
	g1  int
	b0  int
	b1  int
	vol int
}

func getColorIndex(r, g, b int) int {
	return (r << 10) + (r << 6) + r + (g << 5) + g + b
}

/* Histogram is in elements 1..HISTSIZE along each axis,
 * element 0 is for base or marginal value
 * NB: these must start out 0!
 */

// hist3d  build 3-D color histogram of counts, r/g/b, c^2
func hist3d(src [][3]int, size int, vwt, vmr, vmg, vmb *[cubeSize]int, m2 *[cubeSize]float64) []int {
	var i int
	var ind, r, g, b int
	var inr, ing, inb int // index for r,g,b
	var table [256]int
	var qadd []int

	for i = 0; i < 256; i++ {
		table[i] = i * i
	}

	qadd = make([]int, size)
	for i = 0; i < size; i++ {
		r = src[i][0]
		g = src[i][1]
		b = src[i][2]

		inr = (r >> 3) + 1
		ing = (g >> 3) + 1
		inb = (b >> 3) + 1

		ind = getColorIndex(inr, ing, inb)
		vwt[ind]++
		vmr[ind] += r
		vmg[ind] += g
		vmb[ind] += b
		m2[ind] += float64(table[r] + table[g] + table[b])

		qadd[i] = ind
	}
	return qadd
}

/*
  At conclusion of the histogram step, we can interpret
   wt[r][g][b] = sum over voxel of P(c)
   mr[r][g][b] = sum over voxel of r*P(c)  ,  similarly for mg, mb
   m2[r][g][b] = sum over voxel of c^2*P(c)
  Actually each of these should be divided by 'size' to give the usual
  interpretation of P() as ranging from 0 to 1, but we needn't do that here.
*/

/*
  We now convert histogram into moments so that we can rapidly calculate
  the sums of the above quantities over any desired box.
*/

// m3d Compute cumulative moments. */
func m3d(vwt, vmr, vmg, vmb *[cubeSize]int, m2 *[cubeSize]float64) {
	var i, r, g, b int
	var ind1, ind2 int
	var line, lineR, lineG, lineB int
	var line2 float64

	area := [33]int{}
	areaRed := [33]int{}
	areaGreen := [33]int{}
	areaBlue := [33]int{}
	area2 := [33]float64{}

	for r = 1; r <= 32; r++ {
		for i = 0; i <= 32; i++ {
			area[i], areaRed[i], areaGreen[i], areaBlue[i], area2[i] = 0, 0, 0, 0, 0
		}

		for g = 1; g <= 32; g++ {
			line, lineR, lineG, lineB, line2 = 0, 0, 0, 0, 0
			for b = 1; b <= 32; b++ {
				ind1 = getColorIndex(r, g, b)
				line += vwt[ind1]
				lineR += vmr[ind1]
				lineG += vmg[ind1]
				lineB += vmb[ind1]
				line2 += m2[ind1]

				area[b] += line
				areaRed[b] += lineR
				areaGreen[b] += lineG
				areaBlue[b] += lineB
				area2[b] += line2

				ind2 = ind1 - 1089 /* [r-1][g][b] */
				vwt[ind1] = vwt[ind2] + area[b]
				vmr[ind1] = vmr[ind2] + areaRed[b]
				vmg[ind1] = vmg[ind2] + areaGreen[b]
				vmb[ind1] = vmb[ind2] + areaBlue[b]
				m2[ind1] = m2[ind2] + area2[b]
			}
		}
	}
}

// vol Compute sum over a box of any given statistic
func vol(cube *box, moment *[cubeSize]int) int {
	return moment[getColorIndex(cube.r1, cube.g1, cube.b1)] -
		moment[getColorIndex(cube.r1, cube.g1, cube.b0)] -
		moment[getColorIndex(cube.r1, cube.g0, cube.b1)] +
		moment[getColorIndex(cube.r1, cube.g0, cube.b0)] -
		moment[getColorIndex(cube.r0, cube.g1, cube.b1)] +
		moment[getColorIndex(cube.r0, cube.g1, cube.b0)] +
		moment[getColorIndex(cube.r0, cube.g0, cube.b1)] -
		moment[getColorIndex(cube.r0, cube.g0, cube.b0)]
}

// volFloat Computes the volume of the cube in a specific moment. For the floating-point values.
func volFloat(cube *box, moment *[cubeSize]float64) float64 {
	return moment[getColorIndex(cube.r1, cube.g1, cube.b1)] -
		moment[getColorIndex(cube.r1, cube.g1, cube.b0)] -
		moment[getColorIndex(cube.r1, cube.g0, cube.b1)] +
		moment[getColorIndex(cube.r1, cube.g0, cube.b0)] -
		moment[getColorIndex(cube.r0, cube.g1, cube.b1)] +
		moment[getColorIndex(cube.r0, cube.g1, cube.b0)] +
		moment[getColorIndex(cube.r0, cube.g0, cube.b1)] -
		moment[getColorIndex(cube.r0, cube.g0, cube.b0)]
}

/*
  The next two routines allow a slightly more efficient calculation
  of Vol() for a proposed sub box of a given box.  The sum of top()
  and Bottom() is the Vol() of a sub box split in the given direction
  and with the specified new upper bound.
*/

// bottom Compute part of Vol(cube, mmt) that doesn't depend on r1, g1, or b1 (depending on dir)
func bottom(cube *box, direction int, moment *[cubeSize]int) int {
	switch direction {
	case red:
		return -moment[getColorIndex(cube.r0, cube.g1, cube.b1)] +
			moment[getColorIndex(cube.r0, cube.g1, cube.b0)] +
			moment[getColorIndex(cube.r0, cube.g0, cube.b1)] -
			moment[getColorIndex(cube.r0, cube.g0, cube.b0)]
	case green:
		return -moment[getColorIndex(cube.r1, cube.g0, cube.b1)] +
			moment[getColorIndex(cube.r1, cube.g0, cube.b0)] +
			moment[getColorIndex(cube.r0, cube.g0, cube.b1)] -
			moment[getColorIndex(cube.r0, cube.g0, cube.b0)]
	case blue:
		return -moment[getColorIndex(cube.r1, cube.g1, cube.b0)] +
			moment[getColorIndex(cube.r1, cube.g0, cube.b0)] +
			moment[getColorIndex(cube.r0, cube.g1, cube.b0)] -
			moment[getColorIndex(cube.r0, cube.g0, cube.b0)]
	default:
		return 0
	}
}

// top Compute remainder of Vol(cube, mmt), substituting pos for r1, g1, or b1 (depending on dir)
func top(cube *box, direction, position int, moment *[cubeSize]int) int {
	switch direction {
	case red:
		return moment[getColorIndex(position, cube.g1, cube.b1)] -
			moment[getColorIndex(position, cube.g1, cube.b0)] -
			moment[getColorIndex(position, cube.g0, cube.b1)] +
			moment[getColorIndex(position, cube.g0, cube.b0)]
	case green:
		return moment[getColorIndex(cube.r1, position, cube.b1)] -
			moment[getColorIndex(cube.r1, position, cube.b0)] -
			moment[getColorIndex(cube.r0, position, cube.b1)] +
			moment[getColorIndex(cube.r0, position, cube.b0)]

	case blue:
		return moment[getColorIndex(cube.r1, cube.g1, position)] -
			moment[getColorIndex(cube.r1, cube.g0, position)] -
			moment[getColorIndex(cube.r0, cube.g1, position)] +
			moment[getColorIndex(cube.r0, cube.g0, position)]
	default:
		return 0
	}
}

// variance
// Compute the weighted variance of a box
// NB: as with the raw statistics, this is really the variance * size
func variance(cube *box, wt, mr, mg, mb *[cubeSize]int, m2 *[cubeSize]float64) float64 {
	volumeRed := float64(vol(cube, mr))
	volumeGreen := float64(vol(cube, mg))
	volumeBlue := float64(vol(cube, mb))
	volumeMoment := volFloat(cube, m2)
	volumeWeight := float64(vol(cube, wt))

	distance := volumeRed*volumeRed + volumeGreen*volumeGreen + volumeBlue*volumeBlue

	return volumeMoment - (distance / volumeWeight)
}

// maximize
// We want to minimize the sum of the variances of two sub boxes.
// The sum(c^2) terms can be ignored since their sum over both sub boxes
// is the same (the sum for the whole box) no matter where we split.
// The remaining terms have a minus sign in the variance formula,
// so we drop the minus sign and MAXIMIZE the sum of the two terms.
func maximize(cube *box, dir, first, last int, cut *int,
	wholeR, wholeG, wholeB, wholeW int,
	wt, mr, mg, mb *[cubeSize]int) float64 {

	var i int
	var halfR, halfG, halfB, halfW int
	var baseR, baseG, baseB, baseW int
	var temp, max float64

	baseR = bottom(cube, dir, mr)
	baseG = bottom(cube, dir, mg)
	baseB = bottom(cube, dir, mb)
	baseW = bottom(cube, dir, wt)

	max = 0.0
	*cut = -1

	for i = first; i < last; i++ {
		// determines the cube cut at a certain position
		halfR = baseR + top(cube, dir, i, mr)
		halfG = baseG + top(cube, dir, i, mg)
		halfB = baseB + top(cube, dir, i, mb)
		halfW = baseW + top(cube, dir, i, wt)

		/* now half_x is sum over lower half of box, if split at i */
		if halfW == 0 {
			continue // sub box could be empty of pixels!, never split into an empty box
		} else {
			temp = float64(halfR*halfR+halfG*halfG+halfB*halfB) / float64(halfW)
		}

		halfR = wholeR - halfR
		halfG = wholeG - halfG
		halfB = wholeB - halfB
		halfW = wholeW - halfW
		if halfW == 0 {
			continue // sub box could be empty of pixels! Never split into an empty box
		} else {
			temp += float64(halfR*halfR+halfG*halfG+halfB*halfB) / float64(halfW)
		}

		if temp > max {
			max = temp
			*cut = i
		}
	}

	return max
}

func cut(set1, set2 *box, wt, mr, mg, mb *[cubeSize]int) bool {
	var dir int
	var cutR, cutG, cutB int
	var wholeR, wholeG, wholeB, wholeW int
	var maxR, maxG, maxB float64

	wholeR = vol(set1, mr)
	wholeG = vol(set1, mg)
	wholeB = vol(set1, mb)
	wholeW = vol(set1, wt)

	maxR = maximize(set1, red, set1.r0+1, set1.r1, &cutR, wholeR, wholeG, wholeB, wholeW, wt, mr, mg, mb)
	maxG = maximize(set1, green, set1.g0+1, set1.g1, &cutG, wholeR, wholeG, wholeB, wholeW, wt, mr, mg, mb)
	maxB = maximize(set1, blue, set1.b0+1, set1.b1, &cutB, wholeR, wholeG, wholeB, wholeW, wt, mr, mg, mb)

	if (maxR >= maxG) && (maxR >= maxB) {
		dir = red
		if cutR < 0 {
			return false /* can't split the box */
		}
	} else if (maxG >= maxR) && (maxG >= maxB) {
		dir = green
	} else {
		dir = blue
	}

	set2.r1 = set1.r1
	set2.g1 = set1.g1
	set2.b1 = set1.b1

	if dir == red {
		set2.r0, set1.r1 = cutR, cutR
		set2.g0 = set1.g0
		set2.b0 = set1.b0
	} else if dir == green {
		set2.g0, set1.g1 = cutG, cutG
		set2.r0 = set1.r0
		set2.b0 = set1.b0
	} else { /* dir == BLUE */
		set2.b0, set1.b1 = cutB, cutB
		set2.r0 = set1.r0
		set2.g0 = set1.g0
	}

	set1.vol = (set1.r1 - set1.r0) * (set1.g1 - set1.g0) * (set1.b1 - set1.b0)
	set2.vol = (set2.r1 - set2.r0) * (set2.g1 - set2.g0) * (set2.b1 - set2.b0)

	return true
}

func mark(cube *box, label int, tag *[cubeSize]int) {
	var r, g, b int
	for r = cube.r0 + 1; r <= cube.r1; r++ {
		for g = cube.g0 + 1; g <= cube.g1; g++ {
			for b = cube.b0 + 1; b <= cube.b1; b++ {
				tag[getColorIndex(r, g, b)] = label
			}
		}
	}
}

func QuantWu(pixels [][3]int, k int) ([][3]int, error) {
	var lutRgb [maxColor][3]int
	var qadd []int
	var tag [cubeSize]int
	var next int
	var i, j int
	var weight int
	var size int
	var maxColors int
	var wt, mr, mg, mb [cubeSize]int
	var m2 [cubeSize]float64
	var temp float64
	var vv [maxColor]float64
	var cube [maxColor]box
	var count []float64
	var rank []int
	var palettes [][3]int

	maxColors = k

	size = len(pixels)
	qadd = hist3d(pixels, size, &wt, &mr, &mg, &mb, &m2)

	m3d(&wt, &mr, &mg, &mb, &m2)

	cube[0] = box{r1: 32, g1: 32, b1: 32}

	next = 0
	for i = 1; i < maxColors; i++ {
		if cut(&cube[next], &cube[i], &wt, &mr, &mg, &mb) {
			/* Volume test ensures we won't try to cut one-cell box */
			if cube[next].vol > 1 {
				vv[next] = variance(&cube[next], &wt, &mr, &mg, &mb, &m2)
			} else {
				vv[next] = 0
			}

			if cube[i].vol > 1 {
				vv[i] = variance(&cube[i], &wt, &mr, &mg, &mb, &m2)
			} else {
				vv[i] = 0
			}
		} else {
			vv[next] = 0.0 /* Don't try to split this box again */
			i--            /* Didn't create box i */
		}

		next = 0
		temp = vv[0]
		for j = 1; j <= i; j++ {
			if vv[j] > temp {
				temp = vv[j]
				next = j
			}
		}

		if temp <= 0.0 {
			maxColors = i + 1 /* Only got I + 1 boxes */
			break
		}
	}

	for i = 0; i < maxColors; i++ {
		mark(&cube[i], i, &tag)
		weight = vol(&cube[i], &wt)

		if weight > 0 {
			lutRgb[i][0], lutRgb[i][1], lutRgb[i][2] = vol(&cube[i], &mr)/weight, vol(&cube[i], &mg)/weight, vol(&cube[i], &mb)/weight
		} else { /* Bogux box */
			lutRgb[i][0], lutRgb[i][1], lutRgb[i][2] = 0, 0, 0
		}
	}

	count = make([]float64, maxColors)
	for i = 0; i < size; i++ {
		count[tag[qadd[i]]]++
	}

	rank = argsort.Quicksort(count)
	palettes = make([][3]int, k)
	for i = 0; i < maxColors; i++ {
		palettes[i] = lutRgb[rank[maxColors-1-i]]
	}
	return palettes, nil
}
