package main

import (
	"color-thief/helper"
	"color-thief/rgbUtil"
	"log"
	"testing"
)

var (
	p1 [][]int
)

func init() {
	var err error
	img1, err := rgbUtil.ReadImage("../example/photo1.jpg")
	if err != nil {
		log.Fatal(err)
	}
	p1 = helper.SubsamplingPixels(img1)

	if len(p1) != 300*225 {
		log.Fatal("Unexpected sample size found for photo1: ", len(p1))
	}
}

func BenchmarkGetHistogram(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = getHistogram(p1)
	}
}

func BenchmarkWSM(b *testing.B) {
	for i := 0; i < b.N; i++ {
		WSM(p1, 6)
	}
}
