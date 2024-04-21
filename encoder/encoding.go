package main

import "math"

type ImageBlock [16]int

const (
	ENC_SKIP           byte = 0x00
	ENC_SKIP_LONG      byte = 0x10
	ENC_REPEAT         byte = 0x20
	ENC_REPEAT_LONG    byte = 0x30
	ENC_SOLID          byte = 0x40
	ENC_SOLID_LONG     byte = 0x50
	ENC_SOLID_SEP      byte = 0x60
	ENC_SOLID_SEP_LONG byte = 0x70
	ENC_PAL2           byte = 0x80
	ENC_PAL2_CACHE     byte = 0x90
	ENC_PAL4           byte = 0xA0
	ENC_PAL4_CACHE     byte = 0xB0
	ENC_PAL8           byte = 0xC0
	ENC_PAL8_CACHE     byte = 0xD0
	ENC_RAW            byte = 0xE0
	ENC_RAW_LONG       byte = 0xF0
)

type EncodedBlock struct {
	BlockType byte
	Count     int
	MetaData  []int
	PixelData [][]int
}

func ImageToBlocks(image []int, width int, height int) ([]ImageBlock, int, int) {
	bw := int(math.Ceil(float64(width) / 4))
	bh := int(math.Ceil(float64(height) / 4))
	result := make([]ImageBlock, bw*bh)
	for y := 0; y < bh; y++ {
		for x := 0; x < bw; x++ {
			for by := 0; by < 4; by++ {
				for bx := 0; bx < 4; bx++ {
					bi := bx + by*4
					bl := x + y*bw
					imx := x*4 + bx
					if imx >= width {
						imx = width - 1
					}
					imy := y*4 + by
					if imy >= height {
						imy = height - 1
					}
					imi := imx + imy*width
					result[bl][bi] = image[imi]
				}
			}
		}
	}
	return result, bw, bh
}

func BlocksToImage(blocks []ImageBlock, blockWidth int, blockHeight int) ([]int, int, int) {
	width := blockWidth * 4
	height := blockHeight * 4
	result := make([]int, width*height)
	for y := 0; y < blockHeight; y++ {
		for x := 0; x < blockWidth; x++ {
			for by := 0; by < 4; by++ {
				for bx := 0; bx < 4; bx++ {
					bi := bx + by*4
					bl := x + y*blockWidth
					imx := x*4 + bx
					imy := y*4 + by
					imi := imx + imy*width
					result[imi] = blocks[bl][bi]
				}
			}
		}
	}
	return result, width, height
}

//func (block ImageBlock) Compare(other ImageBlock) float64 {}
