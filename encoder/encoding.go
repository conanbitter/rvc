package main

import (
	"fmt"
	"math"
)

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
	Score     float64
	Result    *ImageBlock
}

type FrameEncoder struct {
	chain []EncodedBlock
	pal   Palette
}

//region BLOCKS

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

//endregion

//region ENCODER

func NewEncoder(pal Palette) *FrameEncoder {
	return &FrameEncoder{
		chain: make([]EncodedBlock, 0),
		pal:   pal,
	}
}

func (encoder *FrameEncoder) AddSuggestion(suggestion *EncodeSuggestion) {
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

func (encoder *FrameEncoder) Decode() []ImageBlock {
	result := make([]ImageBlock, 0)
	var block ImageBlock
	var last ImageBlock
	for _, enc := range encoder.chain {
		switch enc.BlockType {
		case ENC_REPEAT:
			result = append(result, last)
		case ENC_SOLID:
			for i := range block {
				block[i] = enc.MetaData[0]
			}
			result = append(result, block)
			last = block
		case ENC_PAL2, ENC_PAL4, ENC_PAL8:
			for i := range block {
				block[i] = enc.MetaData[enc.PixelData[0][i]]
			}
			result = append(result, block)
			last = block
		case ENC_RAW:
			result = append(result, ImageBlock(enc.PixelData[0]))
			last = block
		}
	}
	return result
}

func (encoder *FrameEncoder) DebugDecode() []int {
	result := make([]int, 0)
	for _, enc := range encoder.chain {
		var col int
		switch enc.BlockType {
		case ENC_REPEAT:
			col = 1
		case ENC_SOLID:
			col = 2
		case ENC_PAL2:
			col = 3
		case ENC_PAL4:
			col = 4
		case ENC_PAL8:
			col = 5
		case ENC_RAW:
			col = 6
		}
		col = col | 0b1000
		result = append(result, col)
	}
	return result
}

func (encoder *FrameEncoder) Encode(frame []ImageBlock) {
	encoder.chain = make([]EncodedBlock, 0)
	var counts [6]int
	treshold := float64(0.02)
	var last ImageBlock
	first := true
	for _, block := range frame {
		var suggestion *EncodeSuggestion

		if first {
			first = false
		} else {
			suggestion = ChooseRepeat(&block, &last, encoder.pal)
			if suggestion.Score < treshold {
				encoder.AddSuggestion(suggestion)
				last = *suggestion.Result
				counts[0]++
				continue
			}
		}

		suggestion = ChooseSolid(&block, encoder.pal)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			last = *suggestion.Result
			counts[1]++
			continue
		}

		suggestion = ChooseSubColor(&block, encoder.pal, ENC_PAL2)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			last = *suggestion.Result
			counts[2]++
			continue
		}

		suggestion = ChooseSubColor(&block, encoder.pal, ENC_PAL4)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			last = *suggestion.Result
			counts[3]++
			continue
		}

		suggestion = ChooseSubColor(&block, encoder.pal, ENC_PAL8)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			last = *suggestion.Result
			counts[4]++
			continue
		}

		suggestion = ChooseRaw(&block)
		encoder.AddSuggestion(suggestion)
		last = *suggestion.Result
		counts[5]++
	}
	fmt.Println(len(frame), len(encoder.chain))
	fmt.Printf("  repeat: %d\n  solid:  %d\n  pal2:   %d\n  pal4:   %d\n  pal8:   %d\n  raw:    %d\n", counts[0], counts[1], counts[2], counts[3], counts[4], counts[5])
}

//endregion

//region CHOOSING

func ChooseRaw(source *ImageBlock) *EncodeSuggestion {
	result := make([]int, 16)
	copy(result, (*source)[:])
	return &EncodeSuggestion{
		Encoding:  ENC_RAW,
		MetaData:  nil,
		PixelData: result,
		First:     true,
		Score:     0,
		Result:    source,
	}
}

func ChooseRepeat(source *ImageBlock, last *ImageBlock, pal Palette) *EncodeSuggestion {
	return &EncodeSuggestion{
		Encoding:  ENC_REPEAT,
		MetaData:  nil,
		PixelData: nil,
		First:     true,
		Score:     CompareBlocks(source, last, pal),
		Result:    source,
	}
}

func ChooseSolid(source *ImageBlock, pal Palette) *EncodeSuggestion {
	color := -1
	score := math.MaxFloat64

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
	resultBlock := &ImageBlock{}
	for i := range resultBlock {
		resultBlock[i] = color
	}

	result := &EncodeSuggestion{
		Encoding:  ENC_SOLID,
		MetaData:  []int{color},
		PixelData: nil,
		First:     true,
		Score:     score,
		Result:    resultBlock,
	}

	return result
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

func CalcSubpal(source *ImageBlock, pal Palette, colorNum int) []int {
	calc := NewColorCalcMini(colorNum, 1000, 5)
	calc.Input(source, pal)
	calc.Run()
	return calc.GetSubPal(pal)
}

func encodingToColors(encoding byte) int {
	switch encoding {
	case ENC_PAL2, ENC_PAL2_CACHE:
		return 2
	case ENC_PAL4, ENC_PAL4_CACHE:
		return 4
	case ENC_PAL8, ENC_PAL8_CACHE:
		return 8
	default:
		return 8
	}
}

func ChooseSubColor(source *ImageBlock, pal Palette, encoding byte) *EncodeSuggestion {
	colorNum := encodingToColors(encoding)
	data := CalcSubpal(source, pal, colorNum)
	pixels, resultBlock := ApplySubpal(source, pal, data)
	score := CompareBlocks(source, resultBlock, pal)

	result := &EncodeSuggestion{
		Encoding:  encoding,
		MetaData:  data,
		PixelData: (*pixels)[:],
		First:     true,
		Score:     score,
		Result:    resultBlock,
	}

	return result
}

//endregion
