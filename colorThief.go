package main

import (
	"color-thief/rgbUtil"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"sort"
)

const BucketSize = 64

type SortedColor struct {
	Color string `json:"color"`
	diff  float64
}

type Bucket struct {
	priority int
	value    []uint32
}

type PriorityQueue []*Bucket

func (pq PriorityQueue) Len() int { return len(pq) }

// update modifies the priority and value of an Bucket in the queue.
func (pq *PriorityQueue) update(item *Bucket, value []uint8) {
	item.value[0] += uint32(value[0])
	item.value[1] += uint32(value[1])
	item.value[2] += uint32(value[2])
	item.value[3]++
	item.priority++
}

func GetPalette(src image.Image, count int) []SortedColor {
	pq := make(PriorityQueue, 0)
	m := map[uint32]int{}

	// https://stackoverflow.com/a/59747737
	// convert to 255 scales, four times faster than using At
	bounds := src.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	img := image.NewRGBA(src.Bounds())
	draw.Draw(img, src.Bounds(), src, image.Point{}, draw.Src)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			offset := (y*width + x) * 4
			pix := img.Pix[offset : offset+4]
			index := rgbUtil.GetBucket(pix, BucketSize)
			if val, ok := m[index]; ok {
				pq.update(pq[val], pix)
			} else {
				m[index] = pq.Len()
				pq = append(pq, &Bucket{1, []uint32{uint32(pix[0]), uint32(pix[1]), uint32(pix[2]), 1}})
			}
		}
	}
	sort.Slice(pq, func(i, j int) bool { return pq[i].priority > pq[j].priority })

	dominant := rgbUtil.GetAverageColor(pq[0].value)
	d := SortedColor{Color: rgbUtil.Hex(dominant)}
	if pq.Len() < count {
		h, _, _ := rgbUtil.Hsv(dominant)
		return append([]SortedColor{d}, generatePalette(h, count)...)
	}
	palette := make([]SortedColor, count-1)
	for i, b := range pq[1:count] {
		c := rgbUtil.GetAverageColor(b.value)
		palette[i] = SortedColor{rgbUtil.Hex(c), rgbUtil.DistanceLab(c, dominant)}
	}
	sort.Slice(palette, func(i, j int) bool { return palette[i].diff > palette[j].diff })
	return append([]SortedColor{d}, palette...)
}

func generatePalette(h float64, count int) []SortedColor {
	c := NewColorScheme().FromHue(h).SetScheme("analogic").Variation("soft").SetWebSafe(true).Colors()
	palette := make([]SortedColor, count)
	for i := 0; i < count; i++ {
		palette[i] = SortedColor{Color: c[i]}
	}
	return palette
}
