package main

import (
	"color-thief/helper"
	"color-thief/wsm"
	"color-thief/wu"
	"log"
	"strings"
	"syscall/js"
)

func main() {
	c := make(chan struct{}, 0)

	println("Go WebAssembly Initialized")

	js.Global().Set("getPalette", js.FuncOf(getPalette))

	<-c
}

func getPalette(_ js.Value, args []js.Value) interface{} {
	var pixels, palette [][3]int
	var img []int
	var width, height int
	var k, s int
	var err error
	var sb strings.Builder

	width, height, k, s = args[1].Int(), args[2].Int(), args[3].Int(), args[4].Int()

	img = parsePixelArray(args[0], width, height)

	if len(img) != width*height*4 {
		log.Fatal("invalid image size", len(img))
	}

	pixels = helper.SubsamplingPixels(img, width, height)

	switch s {
	case 0:
		palette, err = wu.QuantWu(pixels, k)
		break
	case 1:
		palette, err = wsm.WSM(pixels, k)
		break
	default:
		log.Fatal("function type should be either 0 or 1")
	}

	if err != nil {
		log.Fatal(err)
	}

	sb = strings.Builder{}
	for i, v := range palette {
		sb.WriteString(helper.Hex(v))
		if i < k-1 {
			sb.WriteString(",")
		}
	}
	return sb.String()
}

func parsePixelArray(arr js.Value, width, height int) []int {
	size := width * height * 4
	pixels := make([]int, size)
	for i := 0; i < size; i++ {
		pixels[i] = arr.Index(i).Int()
	}
	return pixels
}
