package main

import (
	"color-thief/helper"
	"color-thief/rgbUtil"
	"fmt"
	"log"
	"time"
)

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
func hist3d(src [][]int, size int, vwt, vmr, vmg, vmb [][][]int, m2 [][][]float64) []int {
	var Qadd []int // quantized pixels
	var i int
	var count int
	var r, g, b int
	var inr, ing, inb int // index for r,g,b
	var table []int

	table = make([]int, 256)
	for i = 0; i < 256; i++ {
		table[i] = i * i
	}

	Qadd = make([]int, size)
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

		Qadd[count] = getColorIndex(inr, ing, inb)
		count++
	}
	return Qadd
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
	var line, line_r, line_g, line_b int
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
			line, line_r, line_g, line_b, line2 = 0, 0, 0, 0, 0
			for b = 1; b <= 32; b++ {
				line += vwt[r][g][b]
				line_r += vmr[r][g][b]
				line_g += vmg[r][g][b]
				line_b += vmb[r][g][b]
				line2 += m2[r][g][b]

				area[b] += line
				areaRed[b] += line_r
				areaGreen[b] += line_g
				areaBlue[b] += line_b
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

// volumeFloat / <summary>
/// Computes the volume of the cube in a specific moment. For the floating-point values.
/// </summary>
func volumeFloat(cube *box, moment [][][]float64) float64 {
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
  of Vol() for a proposed subbox of a given box.  The sum of top()
  and Bottom() is the Vol() of a subbox split in the given direction
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
	volumeMoment := volumeFloat(cube, m2)
	volumeWeight := float64(vol(cube, wt))

	distance := volumeRed*volumeRed + volumeGreen*volumeGreen + volumeBlue*volumeBlue

	return volumeMoment - (distance / volumeWeight)
}

// maximize
// We want to minimize the sum of the variances of two subboxes.
// The sum(c^2) terms can be ignored since their sum over both subboxes
// is the same (the sum for the whole box) no matter where we split.
// The remaining terms have a minus sign in the variance formula,
// so we drop the minus sign and MAXIMIZE the sum of the two terms.
func maximize(cube *box, dir, first, last int, cut *int,
	whole_r, whole_g, whole_b, whole_w int,
	wt, mr, mg, mb [][][]int) float64 {

	var i int
	var half_r, half_g, half_b, half_w int
	var base_r, base_g, base_b, base_w int
	var temp, max float64

	base_r = bottom(cube, dir, mr)
	base_g = bottom(cube, dir, mg)
	base_b = bottom(cube, dir, mb)
	base_w = bottom(cube, dir, wt)

	max = 0.0
	*cut = -1

	for i = first; i < last; i++ {
		// determines the cube cut at a certain position
		half_r = base_r + top(cube, dir, i, mr)
		half_g = base_g + top(cube, dir, i, mg)
		half_b = base_b + top(cube, dir, i, mb)
		half_w = base_w + top(cube, dir, i, wt)

		/* now half_x is sum over lower half of box, if split at i */
		if half_w == 0 {
			continue // subbox could be empty of pixels!, never split into an empty box
		} else {
			temp = float64(half_r*half_r+half_g*half_g+half_b*half_b) / float64(half_w)
		}

		half_r = whole_r - half_r
		half_g = whole_g - half_g
		half_b = whole_b - half_b
		half_w = whole_w - half_w
		if half_w == 0 {
			continue // Subbox could be empty of pixels! Never split into an empty box
		} else {
			temp += float64(half_r*half_r+half_g*half_g+half_b*half_b) / float64(half_w)
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
	var cutr, cutg, cutb int
	var whole_r, whole_g, whole_b, whole_w int
	var maxr, maxg, maxb float64

	whole_r = vol(set1, mr)
	whole_g = vol(set1, mg)
	whole_b = vol(set1, mb)
	whole_w = vol(set1, wt)

	maxr = maximize(set1, RED, set1.r0+1, set1.r1, &cutr, whole_r, whole_g, whole_b, whole_w, wt, mr, mg, mb)
	maxg = maximize(set1, GREEN, set1.g0+1, set1.g1, &cutg, whole_r, whole_g, whole_b, whole_w, wt, mr, mg, mb)
	maxb = maximize(set1, BLUE, set1.b0+1, set1.b1, &cutb, whole_r, whole_g, whole_b, whole_w, wt, mr, mg, mb)

	if (maxr >= maxg) && (maxr >= maxb) {
		dir = RED
		if cutr < 0 {
			return false /* can't split the box */
		}
	} else if (maxg >= maxr) && (maxg >= maxb) {
		dir = GREEN
	} else {
		dir = BLUE
	}

	set2.r1 = set1.r1
	set2.g1 = set1.g1
	set2.b1 = set1.b1

	if dir == RED {
		set2.r0, set1.r1 = cutr, cutr
		set2.g0 = set1.g0
		set2.b0 = set1.b0
	} else if dir == GREEN {
		set2.g0, set1.g1 = cutg, cutg
		set2.r0 = set1.r0
		set2.b0 = set1.b0
	} else { /* dir == BLUE */
		set2.b0, set1.b1 = cutb, cutb
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
	var lut_rgb [][3]int
	//var tag []int
	//var Qadd []int
	var next int
	var i, j int
	var size int
	var weight int
	var max_colors int
	var wt, mr, mg, mb [][][]int
	var m2 [][][]float64
	var temp float64
	var vv []float64
	var cube []box

	max_colors = k
	size = len(pixels)

	wt = *helper.New3dMatrixInt(33, 33, 33)
	mr = *helper.New3dMatrixInt(33, 33, 33)
	mg = *helper.New3dMatrixInt(33, 33, 33)
	mb = *helper.New3dMatrixInt(33, 33, 33)
	m2 = *helper.New3dMatrixFloat(33, 33, 33)

	_ = hist3d(pixels, size, wt, mr, mg, mb, m2)

	m3d(wt, mr, mg, mb, m2)

	cube = make([]box, MaxColor)
	cube[0] = box{r1: 32, g1: 32, b1: 32}

	next = 0
	vv = make([]float64, MaxColor)
	for i = 1; i < max_colors; i++ {
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
			max_colors = i + 1 /* Only got I + 1 boxes */
			break
		}
	}

	//tag = make([]int, 33*33*33)
	lut_rgb = make([][3]int, MaxColor)

	palette := make([]string, max_colors)
	for i = 0; i < max_colors; i++ {
		//mark(&cube[i], i, tag)
		weight = vol(&cube[i], wt)

		if weight > 0 {
			lut_rgb[i][0] = vol(&cube[i], mr) / weight
			lut_rgb[i][1] = vol(&cube[i], mg) / weight
			lut_rgb[i][2] = vol(&cube[i], mb) / weight
		} else { /* Bogux box */
			lut_rgb[i][0] = 0
			lut_rgb[i][1] = 0
			lut_rgb[i][2] = 0
		}
		palette[i] = rgbUtil.Hex(lut_rgb[i])
	}

	return palette
}

func main() {
	img1, err := rgbUtil.ReadImage("example/photo1.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	p := helper.SubsamplingPixels(img1)
	start := time.Now()
	_ = QuantWu(p, 6)
	fmt.Println(time.Since(start))
}
