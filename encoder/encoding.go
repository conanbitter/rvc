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

type PaletteCache struct {
	Pals  [256][]int
	Count int
	Head  int
}

type FrameEncoder struct {
	chain     []EncodedBlock
	pal       Palette
	lastFrame []ImageBlock
	palcahe   [3]*PaletteCache
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

//region PAL CACHE

func NewPaletteCache() *PaletteCache {
	return &PaletteCache{
		Count: 0,
		Head:  -1,
	}
}

func (palcache *PaletteCache) AddPalette(pal []int) {
	palcache.Head++
	if palcache.Head >= 256 {
		palcache.Head = 0
	}
	palcache.Pals[palcache.Head] = pal
	if palcache.Count < 256 {
		palcache.Count++
	}
}

func (palcache *PaletteCache) Reset() {
	palcache.Count = 0
	palcache.Head = -1
}

func (palcache *PaletteCache) GetPals() [][]int {
	return palcache.Pals[:palcache.Count]
}

//endregion

//region ENCODER

func NewEncoder(pal Palette) *FrameEncoder {
	return &FrameEncoder{
		chain:     make([]EncodedBlock, 0),
		pal:       pal,
		lastFrame: nil,
		palcahe:   [3]*PaletteCache{NewPaletteCache(), NewPaletteCache(), NewPaletteCache()},
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

func (encoder *FrameEncoder) Decode(lastframe []ImageBlock) []ImageBlock {
	result := make([]ImageBlock, 0)
	encoder.palcahe[0].Reset()
	encoder.palcahe[1].Reset()
	encoder.palcahe[2].Reset()
	var block ImageBlock
	var last ImageBlock
	var index = 0
	for _, enc := range encoder.chain {
		switch enc.BlockType {
		case ENC_SKIP:
			result = append(result, lastframe[index])
			last = lastframe[index]
		case ENC_REPEAT:
			result = append(result, last)
		case ENC_SOLID:
			for i := range block {
				block[i] = enc.MetaData[0]
			}
			result = append(result, block)
			last = block
		case ENC_PAL2, ENC_PAL4, ENC_PAL8:
			_, pch := encodingToColors(enc.BlockType)
			encoder.palcahe[pch].AddPalette(enc.MetaData)
			for i := range block {
				block[i] = enc.MetaData[enc.PixelData[0][i]]
			}
			result = append(result, block)
			last = block
		case ENC_PAL2_CACHE, ENC_PAL4_CACHE, ENC_PAL8_CACHE:
			_, pch := encodingToColors(enc.BlockType)
			palcache := encoder.palcahe[pch].Pals[enc.MetaData[0]]
			for i := range block {
				block[i] = palcache[enc.PixelData[0][i]]
			}
			result = append(result, block)
			last = block
		case ENC_RAW:
			result = append(result, ImageBlock(enc.PixelData[0]))
			last = block
		}
		index += enc.Count
	}
	return result
}

func (encoder *FrameEncoder) DebugDecode() []int {
	result := make([]int, 0)
	for _, enc := range encoder.chain {
		var col int
		switch enc.BlockType {
		case ENC_SKIP, ENC_SKIP_LONG:
			col = 0
		case ENC_REPEAT, ENC_REPEAT_LONG:
			col = 1
		case ENC_SOLID, ENC_SOLID_LONG:
			col = 2
		case ENC_SOLID_SEP, ENC_SOLID_SEP_LONG:
			col = 3
		case ENC_PAL2:
			col = 4
		case ENC_PAL2_CACHE:
			col = 5
		case ENC_PAL4:
			col = 6
		case ENC_PAL4_CACHE:
			col = 7
		case ENC_PAL8:
			col = 8
		case ENC_PAL8_CACHE:
			col = 9
		case ENC_RAW, ENC_RAW_LONG:
			col = 10
		}
		col += 11
		result = append(result, col)
	}
	return result
}

func (encoder *FrameEncoder) Encode(frame []ImageBlock) {
	encoder.chain = make([]EncodedBlock, 0)
	encoder.palcahe[0].Reset()
	encoder.palcahe[1].Reset()
	encoder.palcahe[2].Reset()
	var counts [10]int
	treshold := float64(0.02)
	newLastFrame := make([]ImageBlock, len(frame))
	var last ImageBlock
	for i, block := range frame {
		var suggestion *EncodeSuggestion

		if encoder.lastFrame != nil {
			suggestion = ChooseSkip(&block, &encoder.lastFrame[i], encoder.pal)
			if suggestion.Score < treshold {
				encoder.AddSuggestion(suggestion)
				last = *suggestion.Result
				newLastFrame[i] = *suggestion.Result
				counts[0]++
				continue
			}
		}

		if i > 0 {
			suggestion = ChooseRepeat(&block, &last, encoder.pal)
			if suggestion.Score < treshold {
				encoder.AddSuggestion(suggestion)
				last = *suggestion.Result
				newLastFrame[i] = *suggestion.Result
				counts[1]++
				continue
			}
		}

		suggestion = ChooseSolid(&block, encoder.pal)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			last = *suggestion.Result
			newLastFrame[i] = *suggestion.Result
			counts[2]++
			continue
		}

		if encoder.palcahe[0].Count > 0 {
			suggestion = ChooseSubColorCache(&block, encoder.pal, ENC_PAL2_CACHE, encoder.palcahe[0])
			if suggestion.Score < treshold {
				encoder.AddSuggestion(suggestion)
				last = *suggestion.Result
				newLastFrame[i] = *suggestion.Result
				counts[3]++
				continue
			}
		}

		suggestion = ChooseSubColor(&block, encoder.pal, ENC_PAL2)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			encoder.palcahe[0].AddPalette(suggestion.MetaData)
			last = *suggestion.Result
			newLastFrame[i] = *suggestion.Result
			counts[4]++
			continue
		}

		if encoder.palcahe[1].Count > 0 {
			suggestion = ChooseSubColorCache(&block, encoder.pal, ENC_PAL4_CACHE, encoder.palcahe[1])
			if suggestion.Score < treshold {
				encoder.AddSuggestion(suggestion)
				last = *suggestion.Result
				newLastFrame[i] = *suggestion.Result
				counts[5]++
				continue
			}
		}

		suggestion = ChooseSubColor(&block, encoder.pal, ENC_PAL4)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			encoder.palcahe[1].AddPalette(suggestion.MetaData)
			last = *suggestion.Result
			newLastFrame[i] = *suggestion.Result
			counts[6]++
			continue
		}

		if encoder.palcahe[2].Count > 0 {
			suggestion = ChooseSubColorCache(&block, encoder.pal, ENC_PAL8_CACHE, encoder.palcahe[2])
			if suggestion.Score < treshold {
				encoder.AddSuggestion(suggestion)
				last = *suggestion.Result
				newLastFrame[i] = *suggestion.Result
				counts[7]++
				continue
			}
		}

		suggestion = ChooseSubColor(&block, encoder.pal, ENC_PAL8)
		if suggestion.Score < treshold {
			encoder.AddSuggestion(suggestion)
			encoder.palcahe[2].AddPalette(suggestion.MetaData)
			last = *suggestion.Result
			newLastFrame[i] = *suggestion.Result
			counts[8]++
			continue
		}

		suggestion = ChooseRaw(&block)
		encoder.AddSuggestion(suggestion)
		last = *suggestion.Result
		newLastFrame[i] = *suggestion.Result
		counts[9]++
	}
	fmt.Printf("  skip:   %d\n", counts[0])
	fmt.Printf("  repeat: %d\n", counts[1])
	fmt.Printf("  solid:  %d\n", counts[2])
	fmt.Printf("  pal2c:  %d\n", counts[3])
	fmt.Printf("  pal2:   %d\n", counts[4])
	fmt.Printf("  pal4c:  %d\n", counts[5])
	fmt.Printf("  pal4:   %d\n", counts[6])
	fmt.Printf("  pal8c:  %d\n", counts[7])
	fmt.Printf("  pal8:   %d\n", counts[8])
	fmt.Printf("  raw:    %d\n", counts[9])

	encoder.lastFrame = newLastFrame
	//outimg, outw, outh := BlocksToImage(newLastFrame, 80, 60)
	//ImageSave("../data/enctest/test_enc.png", outimg, outw, outh, encoder.pal)
}

//endregion

//region CHOOSING

func ChooseSkip(source *ImageBlock, prev *ImageBlock, pal Palette) *EncodeSuggestion {
	result := *prev
	return &EncodeSuggestion{
		Encoding:  ENC_SKIP,
		MetaData:  nil,
		PixelData: nil,
		First:     true,
		Score:     CompareBlocks(source, prev, pal),
		Result:    &result,
	}
}

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
		Result:    last,
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

func encodingToColors(encoding byte) (colors int, cache int) {
	switch encoding {
	case ENC_PAL2, ENC_PAL2_CACHE:
		return 2, 0
	case ENC_PAL4, ENC_PAL4_CACHE:
		return 4, 1
	case ENC_PAL8, ENC_PAL8_CACHE:
		return 8, 2
	default:
		return 8, 2
	}
}

func ChooseSubColor(source *ImageBlock, pal Palette, encoding byte) *EncodeSuggestion {
	colorNum, _ := encodingToColors(encoding)
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

func ChooseSubColorCache(source *ImageBlock, pal Palette, encoding byte, palcache *PaletteCache) *EncodeSuggestion {
	minscore := math.MaxFloat64
	bestIndex := 0
	var bestPixels *ImageBlock
	var bestResult *ImageBlock
	for i, subpal := range palcache.GetPals() {
		pixels, resultBlock := ApplySubpal(source, pal, subpal)
		score := CompareBlocks(source, resultBlock, pal)
		if score < minscore {
			minscore = score
			bestIndex = i
			bestPixels = pixels
			bestResult = resultBlock
		}
	}

	result := &EncodeSuggestion{
		Encoding:  encoding,
		MetaData:  []int{bestIndex},
		PixelData: (*bestPixels)[:],
		First:     true,
		Score:     minscore,
		Result:    bestResult,
	}

	return result
}

//endregion
