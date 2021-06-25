package helper

import (
	"image"
	"image/draw"
)

// SubsamplingPixels 2.2.1 Implement 2:1 subsampling in the horizontal and vertical directions, so that only
// 1/4-th of the input image pixels are taken into account
func SubsamplingPixels(src image.Image) [][]int {
	bounds := src.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	img := image.NewRGBA(bounds)
	draw.Draw(img, bounds, src, image.Point{}, draw.Src)

	samplingSize := (width/2 + width%2) * (height/2 + height%2)
	pixelArray := New2dMatrixInt(samplingSize, 3)
	idx := 0

	var offset int
	for y := 0; y < height; y += 2 {
		for x := 0; x < width; x += 2 {
			offset = (y*width + x) * 4
			pixelArray[idx][0], pixelArray[idx][1], pixelArray[idx][2] = int(img.Pix[offset]), int(img.Pix[offset+1]), int(img.Pix[offset+2])
			idx++
		}
	}
	return pixelArray
}

func New2dMatrixInt(x, y int) [][]int {
	m := make([][]int, x)
	for i := 0; i < x; i++ {
		m[i] = make([]int, y)
	}
	return m
}

func New2dMatrixFloat(x, y int) [][]float64 {
	m := make([][]float64, x)
	for i := 0; i < x; i++ {
		m[i] = make([]float64, y)
	}
	return m
}

func New3dMatrixInt(x, y, z int) [][][]int {
	m := make([][][]int, x)
	for i := 0; i < x; i++ {
		m[i] = make([][]int, y)
		for j := 0; j < y; j++ {
			m[i][j] = make([]int, z)
		}
	}
	return m
}

func New3dMatrixFloat(x, y, z int) [][][]float64 {
	m := make([][][]float64, x)
	for i := 0; i < x; i++ {
		m[i] = make([][]float64, y)
		for j := 0; j < y; j++ {
			m[i][j] = make([]float64, z)
		}
	}
	return m
}
