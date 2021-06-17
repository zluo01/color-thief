package main

import (
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"os"
)

func main() {
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

	// [{70d8ef 0} {2a1c14 0.8070654904252931} {453a32 0.6789211446939512} {8c8579 0.4494588794287456} {51b5c6 0.12899715626661845} {7ae5f3 0.06724747276084027}]
	expected := []string{"70d8ef", "2a1c14", "453a32", "8c8579", "51b5c6", "7ae5f3"}
	p := GetPalette(img, 6)
	for i, c := range p {
		if c.Color != expected[i] {
			log.Fatal("Unequal color found", c.Color, expected[i])
		}
	}
}
