package helper

import (
	"image"
	"log"
	"testing"
)

var img image.Image

func init() {
	var err error
	img, err = ReadImage("../example/photo1.jpg")
	if err != nil {
		log.Fatal(err)
	}
}

func BenchmarkSubsamplingPixels(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = SubsamplingPixelsFromImage(img)
	}
}

func TestColor(t *testing.T) {
	palettes := [][3]int{
		{108, 206, 225},
		{54, 36, 27},
		{174, 188, 111},
		{107, 112, 99},
		{206, 222, 223},
		{214, 120, 24},
	}
	for _, v := range palettes {
		c := Color(v)
		r, g, b, _ := c.RGBA()
		if int(r>>8) != v[0] || int(g>>8) != v[1] || int(b>>8) != v[2] {
			t.Error("unequal color found", v, c)
		}
	}
}
