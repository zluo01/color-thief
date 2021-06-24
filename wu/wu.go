package wu

import (
	"color-thief/helper"
	"color-thief/rgbUtil"
)

/**********************************************************************
		Go Implementation of Wu's Color Quantizer (v. 2)
			(see Graphics Gems vol. II, pp. 126-133)
	Ported and modified from: https://gist.github.com/bert/1192520
**********************************************************************/

const (
	MaxColor = 256
	RED      = 2
	GREEN    = 1
	BLUE     = 0
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
func hist3d(src [][]int, vwt, vmr, vmg, vmb [][][]int, m2 [][][]float64) {
	var i int
	var count int
	var r, g, b int
	var inr, ing, inb int // index for r,g,b
	var table []int

	table = make([]int, 256)
	for i = 0; i < 256; i++ {
		table[i] = i * i
	}

	for _, v := range src {
		r = v[0]
		g = v[1]
		b = v[2]

		inr = (r >> 3) + 1
		ing = (g >> 3) + 1
		inb = (b >> 3) + 1

		vwt[inr][ing][inb]++
		vmr[inr][ing][inb] += r
		vmg[inr][ing][inb] += g
		vmb[inr][ing][inb] += b
		m2[inr][ing][inb] += float64(table[r] + table[g] + table[b])

		count++
	}
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
func m3d(vwt, vmr, vmg, vmb [][][]int, m2 [][][]float64) {
	var i, r, g, b int
	var line, lineR, lineG, lineB int
	var line2 float64
	area := make([]int, 33)
	areaRed := make([]int, 33)
	areaGreen := make([]int, 33)
	areaBlue := make([]int, 33)
	area2 := make([]float64, 33)

	for r = 1; r <= 32; r++ {
		for i = 0; i <= 32; i++ {
			area[i], areaRed[i], areaGreen[i], areaBlue[i], area2[i] = 0, 0, 0, 0, 0
		}

		for g = 1; g <= 32; g++ {
			line, lineR, lineG, lineB, line2 = 0, 0, 0, 0, 0
			for b = 1; b <= 32; b++ {
				line += vwt[r][g][b]
				lineR += vmr[r][g][b]
				lineG += vmg[r][g][b]
				lineB += vmb[r][g][b]
				line2 += m2[r][g][b]

				area[b] += line
				areaRed[b] += lineR
				areaGreen[b] += lineG
				areaBlue[b] += lineB
				area2[b] += line2

				vwt[r][g][b] = vwt[r-1][g][b] + area[b]
				vmr[r][g][b] = vmr[r-1][g][b] + areaRed[b]
				vmg[r][g][b] = vmg[r-1][g][b] + areaGreen[b]
				vmb[r][g][b] = vmb[r-1][g][b] + areaBlue[b]
				m2[r][g][b] = m2[r-1][g][b] + area2[b]
			}
		}
	}
}

// vol Compute sum over a box of any given statistic
func vol(cube *box, moment [][][]int) int {
	return moment[cube.r1][cube.g1][cube.b1] -
		moment[cube.r1][cube.g1][cube.b0] -
		moment[cube.r1][cube.g0][cube.b1] +
		moment[cube.r1][cube.g0][cube.b0] -
		moment[cube.r0][cube.g1][cube.b1] +
		moment[cube.r0][cube.g1][cube.b0] +
		moment[cube.r0][cube.g0][cube.b1] -
		moment[cube.r0][cube.g0][cube.b0]
}

// volFloat Computes the volume of the cube in a specific moment. For the floating-point values.
func volFloat(cube *box, moment [][][]float64) float64 {
	return moment[cube.r1][cube.g1][cube.b1] -
		moment[cube.r1][cube.g1][cube.b0] -
		moment[cube.r1][cube.g0][cube.b1] +
		moment[cube.r1][cube.g0][cube.b0] -
		moment[cube.r0][cube.g1][cube.b1] +
		moment[cube.r0][cube.g1][cube.b0] +
		moment[cube.r0][cube.g0][cube.b1] -
		moment[cube.r0][cube.g0][cube.b0]
}

/*
  The next two routines allow a slightly more efficient calculation
  of Vol() for a proposed sub box of a given box.  The sum of top()
  and Bottom() is the Vol() of a sub box split in the given direction
  and with the specified new upper bound.
*/

// bottom Compute part of Vol(cube, mmt) that doesn't depend on r1, g1, or b1 (depending on dir)
func bottom(cube *box, direction int, moment [][][]int) int {
	switch direction {
	case RED:
		return -moment[cube.r0][cube.g1][cube.b1] +
			moment[cube.r0][cube.g1][cube.b0] +
			moment[cube.r0][cube.g0][cube.b1] -
			moment[cube.r0][cube.g0][cube.b0]
	case GREEN:
		return -moment[cube.r1][cube.g0][cube.b1] +
			moment[cube.r1][cube.g0][cube.b0] +
			moment[cube.r0][cube.g0][cube.b1] -
			moment[cube.r0][cube.g0][cube.b0]
	case BLUE:
		return -moment[cube.r1][cube.g1][cube.b0] +
			moment[cube.r1][cube.g0][cube.b0] +
			moment[cube.r0][cube.g1][cube.b0] -
			moment[cube.r0][cube.g0][cube.b0]
	default:
		return 0
	}
}

// top Compute remainder of Vol(cube, mmt), substituting pos for r1, g1, or b1 (depending on dir)
func top(cube *box, direction, position int, moment [][][]int) int {
	switch direction {
	case RED:
		return moment[position][cube.g1][cube.b1] -
			moment[position][cube.g1][cube.b0] -
			moment[position][cube.g0][cube.b1] +
			moment[position][cube.g0][cube.b0]
	case GREEN:
		return moment[cube.r1][position][cube.b1] -
			moment[cube.r1][position][cube.b0] -
			moment[cube.r0][position][cube.b1] +
			moment[cube.r0][position][cube.b0]

	case BLUE:
		return moment[cube.r1][cube.g1][position] -
			moment[cube.r1][cube.g0][position] -
			moment[cube.r0][cube.g1][position] +
			moment[cube.r0][cube.g0][position]
	default:
		return 0
	}
}

// variance
// Compute the weighted variance of a box
// NB: as with the raw statistics, this is really the variance * size
func variance(cube *box, wt, mr, mg, mb [][][]int, m2 [][][]float64) float64 {
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
	wt, mr, mg, mb [][][]int) float64 {

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

func cut(set1, set2 *box, wt, mr, mg, mb [][][]int) bool {
	var dir int
	var cutR, cutG, cutB int
	var wholeR, wholeG, wholeB, wholeW int
	var maxR, maxG, maxB float64

	wholeR = vol(set1, mr)
	wholeG = vol(set1, mg)
	wholeB = vol(set1, mb)
	wholeW = vol(set1, wt)

	maxR = maximize(set1, RED, set1.r0+1, set1.r1, &cutR, wholeR, wholeG, wholeB, wholeW, wt, mr, mg, mb)
	maxG = maximize(set1, GREEN, set1.g0+1, set1.g1, &cutG, wholeR, wholeG, wholeB, wholeW, wt, mr, mg, mb)
	maxB = maximize(set1, BLUE, set1.b0+1, set1.b1, &cutB, wholeR, wholeG, wholeB, wholeW, wt, mr, mg, mb)

	if (maxR >= maxG) && (maxR >= maxB) {
		dir = RED
		if cutR < 0 {
			return false /* can't split the box */
		}
	} else if (maxG >= maxR) && (maxG >= maxB) {
		dir = GREEN
	} else {
		dir = BLUE
	}

	set2.r1 = set1.r1
	set2.g1 = set1.g1
	set2.b1 = set1.b1

	if dir == RED {
		set2.r0, set1.r1 = cutR, cutR
		set2.g0 = set1.g0
		set2.b0 = set1.b0
	} else if dir == GREEN {
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

func mark(cube *box, label int, tag []int) {
	for redIndex := cube.r0 + 1; redIndex <= cube.r1; redIndex++ {
		for greenIndex := cube.g0 + 1; greenIndex <= cube.g1; greenIndex++ {
			for blueIndex := cube.b0 + 1; blueIndex <= cube.b1; blueIndex++ {
				tag[(redIndex<<10)+(redIndex<<6)+redIndex+(greenIndex<<5)+greenIndex+blueIndex] = label
			}
		}
	}
}

func QuantWu(pixels [][]int, k int) []string {
	var lutRgb [][3]int
	var next int
	var i, j int
	var weight int
	var maxColors int
	var wt, mr, mg, mb [][][]int
	var m2 [][][]float64
	var temp float64
	var vv []float64
	var cube []box

	maxColors = k

	wt = *helper.New3dMatrixInt(33, 33, 33)
	mr = *helper.New3dMatrixInt(33, 33, 33)
	mg = *helper.New3dMatrixInt(33, 33, 33)
	mb = *helper.New3dMatrixInt(33, 33, 33)
	m2 = *helper.New3dMatrixFloat(33, 33, 33)

	hist3d(pixels, wt, mr, mg, mb, m2)

	m3d(wt, mr, mg, mb, m2)

	cube = make([]box, MaxColor)
	cube[0] = box{r1: 32, g1: 32, b1: 32}

	next = 0
	vv = make([]float64, MaxColor)
	for i = 1; i < maxColors; i++ {
		if cut(&cube[next], &cube[i], wt, mr, mg, mb) {
			/* Volume test ensures we won't try to cut one-cell box */
			if cube[next].vol > 1 {
				vv[next] = variance(&cube[next], wt, mr, mg, mb, m2)
			} else {
				vv[next] = 0
			}

			if cube[i].vol > 1 {
				vv[i] = variance(&cube[i], wt, mr, mg, mb, m2)
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

	lutRgb = make([][3]int, MaxColor)
	palette := make([]string, maxColors)
	for i = 0; i < maxColors; i++ {
		weight = vol(&cube[i], wt)

		if weight > 0 {
			lutRgb[i][0] = vol(&cube[i], mr) / weight
			lutRgb[i][1] = vol(&cube[i], mg) / weight
			lutRgb[i][2] = vol(&cube[i], mb) / weight
		} else { /* Bogux box */
			lutRgb[i][0] = 0
			lutRgb[i][1] = 0
			lutRgb[i][2] = 0
		}
		palette[i] = rgbUtil.Hex(lutRgb[i])
	}

	return palette
}
