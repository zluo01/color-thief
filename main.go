package color_thief

import (
	"color-thief/helper"
	"color-thief/wsm"
	"color-thief/wu"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
)

// GetColorFromFile return the base color from the image file
func GetColorFromFile(imgPath string) (color.Color, error) {
	colors, err := GetPaletteFromFile(imgPath, 10, 0)
	if err != nil {
		return color.RGBA{}, err
	}
	return colors[0], nil
}

// GetColor return the base color from the image
func GetColor(img image.Image, numColors, functionType int) (color.Color, error) {
	colors, err := GetPalette(img, numColors, functionType)
	if err != nil {
		return color.RGBA{}, err
	}
	return colors[0], nil
}

// GetPaletteFromFile return cluster similar colors from the image file
func GetPaletteFromFile(imgPath string, numColors, functionType int) ([]color.Color, error) {
	var img image.Image
	var err error

	// load image
	img, err = helper.ReadImage(imgPath)
	if err != nil {
		return nil, err
	}

	return GetPalette(img, numColors, functionType)
}

// GetPalette return cluster similar colors by the median cut algorithm
func GetPalette(img image.Image, numColors, functionType int) ([]color.Color, error) {
	var palette, pixels [][3]int
	var colors []color.Color

	if numColors < 1 {
		return nil, errors.New("number of colors should be greater than 0")
	}

	pixels = helper.SubsamplingPixelsFromImage(img)
	switch functionType {
	case 0:
		palette = wu.QuantWu(pixels, numColors)
		break
	case 1:
		palette = wsm.WSM(pixels, numColors)
		break
	default:
		return nil, errors.New("function type should be either 0 or 1")
	}

	colors = make([]color.Color, len(palette))
	for i, v := range palette {
		colors[i] = helper.Color(v)
	}
	return colors, nil
}

func PrintColor(colors []color.Color, filename string) error {
	imgWidth := 100 * len(colors)
	imgHeight := 200
	if imgWidth == 0 {
		return errors.New("colors empty")
	}

	palettes := image.NewPaletted(image.Rect(0, 0, imgWidth, imgHeight), colors)

	for x := 0; x < imgWidth; x++ {
		idx := x / 100
		for y := 0; y < imgHeight; y++ {
			palettes.SetColorIndex(x, y, uint8(idx))
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	if err = png.Encode(file, palettes); err != nil {
		return err
	}

	return file.Close()
}
