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
