package main

import (
	"color-thief/helper"
	"color-thief/rgbUtil"
	"log"
	"testing"
)

func BenchmarkGetPalette(b *testing.B) {
	img1, err := rgbUtil.ReadImage("../example/photo1.jpg")
	if err != nil {
		log.Fatalln(err)
	}
	p := helper.SubsamplingPixels(img1)
	for i := 0; i < b.N; i++ {
		_ = QuantWu(p, 6)
	}
}
