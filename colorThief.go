package main

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"net/http"
	"sort"
)

const BucketSize = 64

type ColorWrapper struct {
	r     uint32
	g     uint32
	b     uint32
	count uint32
}

type SortedColor struct {
	Color string  `json:"color"`
	Diff  float64 `json:"diff"`
}

type Bucket struct {
	priority int
	value    ColorWrapper
}

type PriorityQueue []*Bucket

func (pq PriorityQueue) Len() int { return len(pq) }

// update modifies the priority and value of an Bucket in the queue.
func (pq *PriorityQueue) update(item *Bucket, value ColorWrapper) {
	item.value.r += value.r
	item.value.g += value.g
	item.value.b += value.b
	item.value.count++
	item.priority++
}

func NewColor(color color.Color) ColorWrapper {
	r, g, b, _ := color.RGBA()
	return ColorWrapper{r >> 8, g >> 8, b >> 8, 1}
}

func (c ColorWrapper) getBucket() uint32 {
	return (downSample(c.r)*6+downSample(c.g))*6 + downSample(c.b)
}

func downSample(c uint32) uint32 {
	div := c / BucketSize
	if c%BucketSize > BucketSize/2 {
		div += 1
	}
	return div + 1
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
	var D65 = [3]float64{0.95047, 1.00000, 1.08883}
	fy := factorize(y / D65[1])
	l = 1.16*fy - 0.16
	a = 5.0 * (factorize(x/D65[0]) - fy)
	b = 2.0 * (fy - factorize(z/D65[2]))
	return
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

func (c ColorWrapper) getAverageColor() ColorWrapper {
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

func GetPalette(image image.Image, count int) []SortedColor {
	pq := make(PriorityQueue, 0)
	m := map[uint32]int{}
	for i := image.Bounds().Min.X; i < image.Bounds().Max.X; i++ {
		for j := image.Bounds().Min.Y; j < image.Bounds().Max.Y; j++ {
			c := NewColor(image.At(i, j))
			index := c.getBucket()
			if val, ok := m[index]; ok {
				pq.update(pq[val], c)
			} else {
				m[index] = pq.Len()
				bucket := Bucket{1, c}
				pq = append(pq, &bucket)
			}
		}
	}
	sort.Slice(pq, func(i, j int) bool { return pq[i].priority > pq[j].priority })
	size := count
	if size > pq.Len() {
		size = pq.Len()
	}
	var palette = make([]SortedColor, size)
	var dominant ColorWrapper
	for i, b := range pq[:size] {
		c := b.value.getAverageColor()
		if i == 0 {
			dominant = c
		}
		palette[i] = SortedColor{c.Hex(), c.DistanceLab(dominant)}
	}
	sort.Slice(palette, func(i, j int) bool { return palette[i].Diff < palette[j].Diff })
	return palette
}

func LoadImage(url string) (image.Image, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	loadedImage, _, err := image.Decode(res.Body)
	if err != nil {
		return nil, err
	}
	_ = res.Body.Close()
	return loadedImage, nil
}

func Handler(w http.ResponseWriter, r *http.Request) {
	imgUrl := r.URL.Query().Get("img")
	if imgUrl == "" {
		http.Error(w, "Not enough params", http.StatusBadRequest)
		return
	}
	img, err := LoadImage(imgUrl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	palette := GetPalette(img, 6)
	js, err := json.Marshal(palette)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(js)
}
