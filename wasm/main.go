package main

import (
	"color-thief/helper"
	"color-thief/wsm"
	"color-thief/wu"
)

func main() {}

var (
	buffer   []uint8
	palettes []uint8
)

// Function to init our buffer in wasm memory
//export initBuffer
func initBuffer(size int) {
	if len(buffer) == size {
		return
	}
	buffer = make([]uint8, size)
}

// Function to init our palettes in wasm memory
//export initPalettes
func initPalettes(size int) {
	if len(palettes) == size*3 {
		return
	}
	palettes = make([]uint8, size*3)
}

// Function to return a pointer (Index) to our buffer in wasm memory
//export getWasmMemoryBufferPointer
func getWasmMemoryBufferPointer() *uint8 {
	return &buffer[0]
}

// Function to set size for input image
//export getPalettesBufferPointer
func getPalettesBufferPointer() *uint8 {
	return &palettes[0]
}

// Function to return palettes compute from input image
//export getPalette
func getPalette(w, h, k, s int) int {
	if k < 1 || (s != 0 && s != 1) {
		return 0
	}

	var pixels, palette [][3]int

	pixels = helper.SubsamplingPixels(buffer, w, h)

	switch s {
	case 0:
		palette = wu.QuantWu(pixels, k)
		break
	case 1:
		palette = wsm.WSM(pixels, k)
		break
	default:
		return 0
	}

	for i, v := range palette {
		palettes[3*i], palettes[3*i+1], palettes[3*i+2] = uint8(v[0]), uint8(v[1]), uint8(v[2])
	}
	return 1
}
