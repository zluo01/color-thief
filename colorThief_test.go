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
		log.Fatal(err)
	}
	defer func(res *os.File) {
		err := res.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res)
	img, _, err := image.Decode(res)
	if err != nil {
		log.Fatal(err)
	}
	for n := 0; n < b.N; n++ {
		_ = GetPalette(img, 6)
	}
}
