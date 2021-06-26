package helper

import (
	"image"
	"image/draw"
)

// SubsamplingPixels 2.2.1 Implement 2:1 subsampling in the horizontal and vertical directions, so that only
// 1/4-th of the input image pixels are taken into account
func SubsamplingPixels(src image.Image) [][3]int {
	var offset, y, x, idx int
	var samplingSize int
	var pixels [][3]int

	bounds := src.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, src, image.Point{}, draw.Src)

	samplingSize = (width/2 + width%2) * (height/2 + height%2)
	pixels = make([][3]int, samplingSize)

	idx = 0
	for y = 0; y < height; y += 2 {
		for x = 0; x < width; x += 2 {
			offset = (y*width + x) * 4
			pixels[idx][0], pixels[idx][1], pixels[idx][2] = int(img.Pix[offset]), int(img.Pix[offset+1]), int(img.Pix[offset+2])
			idx++
		}
	}
	return pixels
}
