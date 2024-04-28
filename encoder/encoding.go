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

type EncodeSuggestion struct {
	Encoding  byte
	MetaData  []int
	PixelData []int
	First     bool
	Result    ImageBlock
}

type FrameEncoder struct {
	chain []EncodedBlock
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

func CompareBlocks(a *ImageBlock, b *ImageBlock, pal Palette) float64 {
	var acc float64 = 0.0
	for i, col := range a {
		acc += pal[col].ToFloatColor().Difference(pal[b[i]].ToFloatColor())
	}
	return acc
}

func NewEncoder() *FrameEncoder {
	return &FrameEncoder{
		chain: make([]EncodedBlock, 0),
	}
}

func (encoder *FrameEncoder) AddSuggestion(suggestion EncodeSuggestion) {
	if suggestion.First {
		encoder.chain = append(encoder.chain, EncodedBlock{
			BlockType: suggestion.Encoding,
			Count:     1,
			MetaData:  suggestion.MetaData,
			PixelData: [][]int{suggestion.PixelData},
		})
	} else {
		lastElement := &encoder.chain[len(encoder.chain)-1]
		lastElement.PixelData = append(lastElement.PixelData, suggestion.PixelData)
	}
}

func ChooseSolid(source *ImageBlock, pal Palette) (color int, score float64, result *ImageBlock) {
	color = -1
	score = math.MaxFloat64

	for i, palcolor := range pal {
		var scoreacc float64 = 0
		for _, pixel := range source {
			scoreacc += palcolor.ToFloatColor().Difference(pal[pixel].ToFloatColor())
		}
		if scoreacc < score {
			color = i
			score = scoreacc
		}
	}
	result = &ImageBlock{}
	for i, _ := range result {
		result[i] = color
	}
	return
}

func chooseColor(source int, pal Palette, subpal []int) int {
	best := math.MaxFloat64
	result := 0
	for i, index := range subpal {
		dist := pal[source].ToFloatColor().Difference(pal[index].ToFloatColor())
		if dist < best {
			result = i
			best = dist
		}
	}
	return result
}

func ApplySubpal(source *ImageBlock, pal Palette, subpal []int) (data *ImageBlock, result *ImageBlock) {
	data = &ImageBlock{}
	result = &ImageBlock{}

	for i, color := range source {
		newColor := chooseColor(color, pal, subpal)
		data[i] = newColor
		result[i] = subpal[newColor]
	}
	return
}

//func CalcSubpal(source *ImageBlock, pal Palette, colorNum int) []int {}
