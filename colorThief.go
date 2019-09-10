package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"sort"
)

const BucketSize = 64

type SortedColor struct {
	Color string `json:"color"`
	rgb   ColorWrapper
	diff  float64
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

func GetPalette(image image.Image, count int) []SortedColor {
	pq := make(PriorityQueue, 0)
	m := map[uint32]int{}
	for i := image.Bounds().Min.X; i < image.Bounds().Max.X; i++ {
		for j := image.Bounds().Min.Y; j < image.Bounds().Max.Y; j++ {
			c := NewColor(image.At(i, j))
			index := c.GetBucket()
			if val, ok := m[index]; ok {
				pq.update(pq[val], c)
			} else {
				m[index] = pq.Len()
				pq = append(pq, &Bucket{1, c})
			}
		}
	}
	sort.Slice(pq, func(i, j int) bool { return pq[i].priority > pq[j].priority })

	if pq.Len() < count {
		h, _, _ := pq[0].value.Hsv()
		return generatePalette(h, count)
	}
	var palette = make([]SortedColor, count)
	var dominant ColorWrapper
	for i, b := range pq[:count] {
		c := b.value.GetAverageColor()
		if i == 0 {
			dominant = c
		}
		palette[i] = SortedColor{c.Hex(), c, c.DistanceLab(dominant)}
	}
	sort.Slice(palette, func(i, j int) bool { return palette[i].diff < palette[j].diff })
	return palette
}

func generatePalette(h float64, count int) []SortedColor {
	c := NewColorScheme().FromHue(h).SetScheme("analogic").Variation("soft").SetWebSafe(true).Colors()
	var palette = make([]SortedColor, count)
	for i := 0; i < count; i++ {
		palette[i] = SortedColor{Color: c[i]}
	}
	return palette
}
