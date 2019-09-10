package main

import (
	"log"
	"math"
	"strconv"
)

var SCHEME = map[string]struct{}{
	"mono":          {},
	"monochromatic": {},
	"contrast":      {},
	"triade":        {},
	"tetrade":       {},
	"analogic":      {},
}

var PRESET = map[string][]float64{
	"default": {-1, -1, 1, -0.7, 0.25, 1, 0.5, 1},
	"pastel":  {0.5, -0.9, 0.5, 0.5, 0.1, 0.9, 0.75, 0.75},
	"soft":    {0.3, -0.8, 0.3, 0.5, 0.1, 0.9, 0.5, 0.75},
	"light":   {0.25, 1, 0.5, 0.75, 0.1, 1, 0.5, 1},
	"hard":    {1, -1, 1, -0.6, 0.1, 1, 0.6, 1},
	"pale":    {0.1, -0.85, 0.1, 0.5, 0.1, 1, 0.1, 0.75},
}

var ColorWheel = map[int][]float64{
	0:   {255, 0, 0, 100},
	15:  {255, 51, 0, 100},
	30:  {255, 102, 0, 100},
	45:  {255, 128, 0, 100},
	60:  {255, 153, 0, 100},
	75:  {255, 178, 0, 100},
	90:  {255, 204, 0, 100},
	105: {255, 229, 0, 100},
	120: {255, 255, 0, 100},
	135: {204, 255, 0, 100},
	150: {153, 255, 0, 100},
	165: {51, 255, 0, 100},
	180: {0, 204, 0, 80},
	195: {0, 178, 102, 70},
	210: {0, 153, 153, 60},
	225: {0, 102, 178, 70},
	240: {0, 51, 204, 80},
	255: {25, 25, 178, 70},
	270: {51, 0, 153, 60},
	285: {64, 0, 153, 60},
	300: {102, 0, 153, 60},
	315: {153, 0, 153, 60},
	330: {204, 0, 153, 80},
	345: {229, 0, 102, 90},
}

type ColorScheme struct {
	Color      []MutableColor
	Scheme     string
	Distance   float64
	WebSafe    bool
	Complement bool
}

func NewColorScheme() *ColorScheme {
	colors := make([]MutableColor, 4)
	for i := 0; i < 4; i++ {
		colors[i] = NewMutableColor(60)
	}
	return &ColorScheme{
		Color:      colors,
		Scheme:     "mono",
		Distance:   0.5,
		WebSafe:    false,
		Complement: false,
	}
}

func (c *ColorScheme) SetScheme(s string) *ColorScheme {
	if _, ok := SCHEME[s]; !ok {
		log.Fatalf("'%s' isn't a valid scheme name", s)
	}
	c.Scheme = s
	return c
}

func (c *ColorScheme) SetDistance(d float64) *ColorScheme {
	if d < 0 || d > 1 {
		log.Fatalf("distance %f - argument must be within range [0, 1]", d)
	}
	c.Distance = d
	return c
}

func (c *ColorScheme) SetWebSafe(b bool) *ColorScheme {
	c.WebSafe = b
	return c
}

func (c *ColorScheme) AddComplement(b bool) *ColorScheme {
	c.Complement = b
	return c
}

func (c *ColorScheme) Colors() map[int]string {
	usedColors := 1
	h := c.Color[0].hue
	switch c.Scheme {
	case "mono":
		break
	case "monochromatic":
		break
	case "contrast":
		usedColors = 2
		c.Color[1].setHue(h)
		c.Color[1].rotate(180)
		break
	case "triade":
		usedColors = 3
		dif := 60 * c.Distance
		c.Color[1].setHue(h)
		c.Color[1].rotate(180 - dif)
		c.Color[2].setHue(h)
		c.Color[2].rotate(180 - dif)
		break
	case "tetrade":
		usedColors = 4
		dif := 90 * c.Distance
		c.Color[1].setHue(h)
		c.Color[1].rotate(180)
		c.Color[2].setHue(h)
		c.Color[2].rotate(180 + dif)
		c.Color[3].setHue(h)
		c.Color[3].rotate(dif)
		break
	case "analogic":
		usedColors = 3
		if c.Complement {
			usedColors = 4
		}
		dif := 60 * c.Distance
		c.Color[1].setHue(h)
		c.Color[1].rotate(dif)
		c.Color[2].setHue(h)
		c.Color[2].rotate(360 - dif)
		c.Color[3].setHue(h)
		c.Color[3].rotate(180)
		break
	default:
		log.Fatalf("Unknown color scheme name: %s", c.Scheme)
	}

	ref := usedColors - 1
	output := map[int]string{}
	if ref >= 0 {
		for i := 0; i <= ref; i++ {
			for j := 0; j <= 3; j++ {
				output[i*4+j] = c.Color[i].getHex(c.WebSafe, j)
			}
		}
	} else {
		for i := 0; i >= ref; i-- {
			for j := 0; j <= 3; j++ {
				output[i*4+j] = c.Color[i].getHex(c.WebSafe, j)
			}
		}
	}
	return output
}

func (c *ColorScheme) FromHue(h float64) *ColorScheme {
	c.Color[0].setHue(h)
	return c
}

// 'name' must be a valid color variation name. See "Color Variations"
func (c *ColorScheme) Variation(v string) *ColorScheme {
	if val, ok := PRESET[v]; ok {
		c.setVariantPreset(val)
		return c
	}
	log.Fatalf("'%s' isn't a valid variation name", v)
	return nil
}

func (c *ColorScheme) setVariantPreset(p []float64) [][]float64 {
	results := make([][]float64, 4)
	for i := 0; i <= 3; i++ {
		results[i] = c.Color[i].setVariantPreset(p)
	}
	return results
}

type MutableColor struct {
	hue            float64
	saturation     []float64
	value          []float64
	baseRed        float64
	baseGreen      float64
	baseBlue       float64
	baseSaturation float64
	baseValue      float64
}

func NewMutableColor(hue float64) MutableColor {
	m := MutableColor{saturation: make([]float64, 4), value: make([]float64, 4)}
	m.setHue(hue)
	m.setVariantPreset(PRESET["default"])
	return m
}

func avrg(a, b, k float64) float64 {
	return a + math.Round((b-a)*k)
}

func (m *MutableColor) setHue(h float64) float64 {
	m.hue = math.Mod(h, 360)
	d := math.Mod(m.hue, 15)
	k := d / 15
	d1 := m.hue - d
	d2 := math.Mod(d1+15, 360)
	if d1 == 360 {
		d1 = 0
	}
	if d2 == 360 {
		d2 = 0
	}
	c1 := ColorWheel[int(d1)]
	c2 := ColorWheel[int(d2)]

	m.baseRed = avrg(c1[0], c2[0], k)
	m.baseGreen = avrg(c1[1], c2[1], k)
	m.baseBlue = avrg(c1[2], c2[2], k)
	m.baseValue = avrg(c1[3], c2[3], k)
	m.baseSaturation = avrg(100, 100, k) / 100
	m.baseValue /= 100
	return m.baseValue
}

func (m MutableColor) getHex(webSafe bool, variation int) string {
	max := math.Max(m.baseRed, math.Max(m.baseGreen, m.baseBlue))
	var v = m.baseValue
	var s = m.baseSaturation
	var k float64
	if variation >= 0 {
		v = m.getValue(variation)
		s = m.getSaturation(variation)
	}
	v *= 255
	if max > 0 {
		k = v / max
	}
	rgb := [3]float64{
		math.Min(255, math.Round(v-(v-m.baseRed*k)*s)),
		math.Min(255, math.Round(v-(v-m.baseGreen*k)*s)),
		math.Min(255, math.Round(v-(v-m.baseBlue*k)*s)),
	}
	if webSafe {
		for i, v := range rgb {
			rgb[i] = math.Round(v/51) * 51
		}
	}
	var formatted string
	for i := 0; i < len(rgb); i++ {
		str := strconv.FormatInt(int64(rgb[i]), 16)
		if len(str) < 2 {
			str = "0" + str
		}
		formatted += str
	}
	return formatted
}

func (m *MutableColor) rotate(angle float64) {
	m.setHue(math.Mod(m.hue+angle, 360))
}

func (m MutableColor) getSaturation(variation int) float64 {
	x := m.saturation[variation]
	s := x
	if x < 0 {
		s = -x * m.baseSaturation
	}
	if s > 1 {
		s = 1
	}
	if s < 0 {
		s = 0
	}
	return s
}

func (m MutableColor) getValue(variation int) float64 {
	x := m.value[variation]
	v := x
	if x < 0 {
		v = -x * m.baseValue
	}
	if v > 1 {
		v = 1
	}
	if v < 0 {
		v = 0
	}
	return v
}

func (m *MutableColor) setVariantPreset(p []float64) []float64 {
	results := make([]float64, 4)
	for i := 0; i <= 3; i++ {
		m.saturation[i] = p[2*i]
		m.value[i] = p[2*i+1]
		results[i] = p[2*i+1]
	}
	return results
}
