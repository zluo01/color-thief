package rgbUtil

import (
	"fmt"
	"math"
)

func GetBucket(c []uint8, bucketSize uint32) uint32 {
	return (downSample(uint32(c[0]), bucketSize)*6+downSample(uint32(c[1]), bucketSize))*6 + downSample(uint32(c[2]), bucketSize)
}

func Hsv(c []uint32) (h, s, v float64) {
	r := float64(c[0]) / 255.0
	g := float64(c[1]) / 255.0
	b := float64(c[2]) / 255.0
	min := math.Min(math.Min(r, g), b)
	v = math.Max(math.Max(r, g), b)
	C := v - min

	s = 0.0
	if v != 0.0 {
		s = C / v
	}

	h = 0.0 // We use 0 instead of undefined as in wp.
	if min != v {
		if v == r {
			h = math.Mod((g-b)/C, 6.0)
		}
		if v == g {
			h = (b-r)/C + 2.0
		}
		if v == b {
			h = (r-g)/C + 4.0
		}
		h *= 60.0
		if h < 0.0 {
			h += 360.0
		}
	}
	return
}

func Hex(c []uint32) string {
	return fmt.Sprintf("%02x%02x%02x", uint8(c[0]), uint8(c[1]), uint8(c[2]))
}

func Lab(c []uint32) (l, a, b float64) {
	ri := linearized(float64(c[0]) / 255.0)
	gi := linearized(float64(c[1]) / 255.0)
	bi := linearized(float64(c[2]) / 255.0)

	// rgb to xyz
	x := 0.4124564*ri + 0.3575761*gi + 0.1804375*bi
	y := 0.2126729*ri + 0.7151522*gi + 0.0721750*bi
	z := 0.0193339*ri + 0.1191920*gi + 0.9503041*bi

	// xyz to lab
	D65 := [3]float64{0.95047, 1.00000, 1.08883}
	fy := factorize(y / D65[1])
	l = 1.16*fy - 0.16
	a = 5.0 * (factorize(x/D65[0]) - fy)
	b = 2.0 * (fy - factorize(z/D65[2]))
	return
}

func GetAverageColor(c []uint32) []uint32 {
	if c[3] == 0 {
		return c
	}
	return []uint32{c[0] / c[3], c[1] / c[3], c[2] / c[3]}
}

func DistanceLab(src, dst []uint32) float64 {
	l1, a1, b1 := Lab(src)
	l2, a2, b2 := Lab(dst)
	return math.Sqrt(math.Pow(l1-l2, 2) + math.Pow(a1-a2, 2) + math.Pow(b1-b2, 2))
}

func downSample(c, bucketSize uint32) uint32 {
	div := c / bucketSize
	if c%bucketSize > bucketSize/2 {
		div += 1
	}
	return div + 1
}

func linearized(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func factorize(t float64) float64 {
	if t > 6.0/29.0*6.0/29.0*6.0/29.0 {
		return math.Cbrt(t)
	}
	return t/3.0*29.0/6.0*29.0/6.0 + 4.0/29.0
}
