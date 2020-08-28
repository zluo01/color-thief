package main

import (
	"fmt"
	"image/color"
	"math"
)

// range [0, 255]
type ColorWrapper struct {
	r     uint32
	g     uint32
	b     uint32
	count uint32
}

func NewColor(color color.Color) ColorWrapper {
	r, g, b, _ := color.RGBA()
	return ColorWrapper{r >> 8, g >> 8, b >> 8, 1}
}

func (c ColorWrapper) GetBucket() uint32 {
	return (downSample(c.r)*6+downSample(c.g))*6 + downSample(c.b)
}

func (c ColorWrapper) Hsv() (h, s, v float64) {
	r := float64(c.r) / 255.0
	g := float64(c.g) / 255.0
	b := float64(c.b) / 255.0
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

func (c ColorWrapper) Hex() string {
	return fmt.Sprintf("%02x%02x%02x", uint8(c.r), uint8(c.g), uint8(c.b))
}

func (c ColorWrapper) Lab() (l, a, b float64) {
	ri := linearized(float64(c.r) / 255.0)
	gi := linearized(float64(c.g) / 255.0)
	bi := linearized(float64(c.b) / 255.0)

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

func (c ColorWrapper) GetAverageColor() ColorWrapper {
	if c.count == 0 {
		return c
	}
	return ColorWrapper{r: c.r / c.count, g: c.g / c.count, b: c.b / c.count}
}

func (c ColorWrapper) DistanceLab(color ColorWrapper) float64 {
	l1, a1, b1 := c.Lab()
	l2, a2, b2 := color.Lab()
	return math.Sqrt(math.Pow(l1-l2, 2) + math.Pow(a1-a2, 2) + math.Pow(b1-b2, 2))
}

func downSample(c uint32) uint32 {
	div := c / BucketSize
	if c%BucketSize > BucketSize/2 {
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
