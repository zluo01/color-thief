package helper

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"os"
)

// SubsamplingPixels 2.2.1 Implement 2:1 subsampling in the horizontal and vertical directions, so that only
// 1/4-th of the input image pixels are taken into account
func SubsamplingPixels(src []uint8, width, height int) [][3]int {
	var offset, y, x, idx int
	var samplingSize int
	var pixels [][3]int

	samplingSize = (width/2 + width%2) * (height/2 + height%2)
	pixels = make([][3]int, samplingSize)

	idx = 0
	for y = 0; y < height; y += 2 {
		for x = 0; x < width; x += 2 {
			offset = (y*width + x) * 4
			pixels[idx][0], pixels[idx][1], pixels[idx][2] = int(src[offset]), int(src[offset+1]), int(src[offset+2])
			idx++
		}
	}
	return pixels
}

// SubsamplingPixelsFromImage 2.2.1 Implement 2:1 subsampling in the horizontal and vertical directions, so that only
// 1/4-th of the input image pixels are taken into account
func SubsamplingPixelsFromImage(src image.Image) [][3]int {
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

func Hex(c [3]int) string {
	return fmt.Sprintf("#%02x%02x%02x", uint8(c[0]), uint8(c[1]), uint8(c[2]))
}

func Color(c [3]int) color.Color {
	return color.RGBA{
		R: uint8(c[0]),
		G: uint8(c[1]),
		B: uint8(c[2]),
		A: 255,
	}
}

func ReadImage(uri string) (image.Image, error) {
	res, err := os.Open(uri)
	if err != nil {
		return nil, err
	}

	img, _, err := image.Decode(res)
	if err != nil {
		return nil, err
	}

	if err = res.Close(); err != nil {
		return nil, err
	}
	return img, nil
}
