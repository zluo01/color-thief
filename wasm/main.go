package main

import (
	"color-thief/helper"
	"color-thief/wsm"
	"color-thief/wu"
	"fmt"
	"strings"
)

func main() {}

var buffer []uint8

// Function to return a pointer (Index) to our buffer in wasm memory
//export initBuffer
func initBuffer(size int) {
	buffer = make([]uint8, size)
}

// Function to return a pointer (Index) to our buffer in wasm memory
//export getWasmMemoryBufferPointer
func getWasmMemoryBufferPointer() *uint8 {
	return &buffer[0]
}

// Function to return palettes compute from our buffer in wasm memory
//export getPalette
func getPalette(width, height, k, s int) string {
	var pixels, palette [][3]int
	var sb strings.Builder

	pixels = helper.SubsamplingPixels(buffer, width, height)

	switch s {
	case 0:
		palette = wu.QuantWu(pixels, k)
		break
	case 1:
		palette = wsm.WSM(pixels, k)
		break
	default:
		fmt.Println("function type should be either 0 or 1") // Todo fix later
		return ""
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
