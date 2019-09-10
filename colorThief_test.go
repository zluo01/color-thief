package main

import (
	"image"
	"log"
	"os"
	"testing"
)

func BenchmarkGetPalette(b *testing.B) {
	res, err := os.Open("example/photo1.jpg")
	if err != nil {
		log.Fatalln(err.Error())
	}
	defer res.Close()
	img, _, err := image.Decode(res)
	if err != nil {
		log.Fatalln(err.Error())
	}
	for n := 0; n < b.N; n++ {
		GetPalette(img, 6)
	}
}
