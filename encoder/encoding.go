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
	palcache  [3]*PaletteCache
}

type Chooser func(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion

var DecisionGraph = [][]Chooser{
	{ChooseSkipCont, ChooseRepeatCont, ChooseSolidCont},    // Tier 1 (0 bytes)
	{ChooseSkip, ChooseRepeat},                             // Tier 2 (1 byte)
	{ChooseSolid, ChoosePal2Cont, ChoosePal2CacheCont},     // Tier 3 (2 bytes)
	{ChoosePal2Cache, ChoosePal4Cont, ChoosePal4CacheCont}, // Tier 4 (4 bytes)
	{ChoosePal2}, // Tier 5 (5 bytes)
	{ChoosePal4Cache, ChoosePal8Cont, ChoosePal8CacheCont}, // Tier 6 (6 bytes)
	{ChoosePal8Cache}, // Tier 7 (8 bytes)
	{ChoosePal4},      // Tier 8 (9 bytes)
	{ChoosePal8},      // Tier 9 (15 bytes)
	//{ChooseRaw},       // Tier 10 (16-17 bytes)
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
		palcache:  [3]*PaletteCache{NewPaletteCache(), NewPaletteCache(), NewPaletteCache()},
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
		lastElement.Count++
	}
}

func (encoder *FrameEncoder) GetLastSuggestion() *EncodedBlock {
	if len(encoder.chain) == 0 {
		return nil
	}
	return &encoder.chain[len(encoder.chain)-1]
}

func (encoder *FrameEncoder) Decode(lastframe []ImageBlock) []ImageBlock {
	result := make([]ImageBlock, 0)
	encoder.palcache[0].Reset()
	encoder.palcache[1].Reset()
	encoder.palcache[2].Reset()
	var block ImageBlock
	var last ImageBlock
	var index = 0
	for _, enc := range encoder.chain {
		switch enc.BlockType {
		case ENC_SKIP:
			for i := 0; i < enc.Count; i++ {
				result = append(result, lastframe[index])
				last = lastframe[index]
				index++
			}
		case ENC_REPEAT:
			for i := 0; i < enc.Count; i++ {
				result = append(result, last)
				index++
			}
		case ENC_SOLID:
			for i := range block {
				block[i] = enc.MetaData[0]
			}
			for i := 0; i < enc.Count; i++ {
				result = append(result, block)
				index++
			}
			last = block
		case ENC_PAL2, ENC_PAL4, ENC_PAL8:
			_, pch := encodingToColors(enc.BlockType)
			encoder.palcache[pch].AddPalette(enc.MetaData)
			for _, data := range enc.PixelData {
				for i := range block {
					block[i] = enc.MetaData[data[i]]
				}
				result = append(result, block)
				index++
			}
			last = block
		case ENC_PAL2_CACHE, ENC_PAL4_CACHE, ENC_PAL8_CACHE:
			_, pch := encodingToColors(enc.BlockType)
			palcache := encoder.palcache[pch].Pals[enc.MetaData[0]]
			for _, data := range enc.PixelData {
				for i := range block {
					block[i] = palcache[data[i]]
				}
				result = append(result, block)
				index++
			}
			last = block
		case ENC_RAW:
			for _, data := range enc.PixelData {
				result = append(result, ImageBlock(data))
				index++
			}
			last = ImageBlock(enc.PixelData[len(enc.PixelData)-1])
		}
		//index += enc.Count
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
		result = append(result, col+11)
		for i := 1; i < enc.Count; i++ {
			result = append(result, col)
		}
	}
	return result
}

func ChooseEncoding(input *ImageBlock, treshold float64, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	var suggestion *EncodeSuggestion
	for _, tier := range DecisionGraph {
		suggestion = nil
		for _, chooser := range tier {
			newSuggestion := chooser(input, prev, index, encoder)
			if suggestion == nil || (newSuggestion != nil && newSuggestion.Score < suggestion.Score) {
				suggestion = newSuggestion
			}
		}
		if suggestion != nil && suggestion.Score < treshold {
			return suggestion
		}
	}
	return ChooseRaw(input, prev, index, encoder)
}

func (encoder *FrameEncoder) Encode(frame []ImageBlock) {
	encoder.chain = make([]EncodedBlock, 0)
	encoder.palcache[0].Reset()
	encoder.palcache[1].Reset()
	encoder.palcache[2].Reset()
	counts := make(map[byte]int)
	treshold := float64(0.02)
	newLastFrame := make([]ImageBlock, len(frame))
	var last ImageBlock
	for i, block := range frame {
		suggestion := ChooseEncoding(&block, treshold, &last, i, encoder)
		encoder.AddSuggestion(suggestion)
		last = *suggestion.Result
		newLastFrame[i] = *suggestion.Result
		counts[suggestion.Encoding] = counts[suggestion.Encoding] + 1
		if suggestion.Encoding == ENC_PAL2 && suggestion.First {
			encoder.palcache[0].AddPalette(suggestion.MetaData)
		}
		if suggestion.Encoding == ENC_PAL4 && suggestion.First {
			encoder.palcache[1].AddPalette(suggestion.MetaData)
		}
		if suggestion.Encoding == ENC_PAL8 && suggestion.First {
			encoder.palcache[2].AddPalette(suggestion.MetaData)
		}
	}
	fmt.Printf("  skip:   %d\n", counts[ENC_SKIP])
	fmt.Printf("  repeat: %d\n", counts[ENC_REPEAT])
	fmt.Printf("  solid:  %d\n", counts[ENC_SOLID])
	fmt.Printf("  pal2:   %d\n", counts[ENC_PAL2])
	fmt.Printf("  pal2c:  %d\n", counts[ENC_PAL2_CACHE])
	fmt.Printf("  pal4:   %d\n", counts[ENC_PAL4])
	fmt.Printf("  pal4c:  %d\n", counts[ENC_PAL4_CACHE])
	fmt.Printf("  pal8:   %d\n", counts[ENC_PAL8])
	fmt.Printf("  pal8c:  %d\n", counts[ENC_PAL8_CACHE])
	fmt.Printf("  raw:    %d\n", counts[ENC_RAW])

	encoder.lastFrame = newLastFrame
	//outimg, outw, outh := BlocksToImage(newLastFrame, 80, 60)
	//ImageSave("../data/enctest/test_enc.png", outimg, outw, outh, encoder.pal)
	fmt.Println(len(encoder.chain), len(frame))
}

func (encoder *FrameEncoder) GetFrameSize() int {
	result := 0
	for _, enc := range encoder.chain {
		switch enc.BlockType {
		case ENC_SKIP:
			result += 1
		case ENC_SKIP_LONG:
			result += 2
		case ENC_REPEAT:
			result += 1
		case ENC_REPEAT_LONG:
			result += 2
		case ENC_SOLID:
			result += 2
		case ENC_SOLID_LONG:
			result += 3
		case ENC_SOLID_SEP:
			result += 1 + enc.Count
		case ENC_SOLID_SEP_LONG:
			result += 2 + enc.Count
		case ENC_PAL2:
			result += 1 + 2 + enc.Count*2
		case ENC_PAL2_CACHE:
			result += 1 + 1 + enc.Count*2
		case ENC_PAL4:
			result += 1 + 4 + enc.Count*4
		case ENC_PAL4_CACHE:
			result += 1 + 1 + enc.Count*4
		case ENC_PAL8:
			result += 1 + 8 + enc.Count*6
		case ENC_PAL8_CACHE:
			result += 1 + 1 + enc.Count*6
		case ENC_RAW:
			result += 1 + enc.Count*16
		case ENC_RAW_LONG:
			result += 2 + enc.Count*16
		}
	}
	return result
}

//endregion

//region ENCODE SUGGESTIONS

func SuggestSkip(source *ImageBlock, index int, encoder *FrameEncoder, cont bool) *EncodeSuggestion {
	result := encoder.lastFrame[index]
	return &EncodeSuggestion{
		Encoding:  ENC_SKIP,
		MetaData:  nil,
		PixelData: nil,
		First:     !cont,
		Score:     CompareBlocks(source, &result, encoder.pal),
		Result:    &result,
	}
}

func SuggestRaw(source *ImageBlock, cont bool) *EncodeSuggestion {
	result := make([]int, 16)
	copy(result, (*source)[:])
	return &EncodeSuggestion{
		Encoding:  ENC_RAW,
		MetaData:  nil,
		PixelData: result,
		First:     !cont,
		Score:     0,
		Result:    source,
	}
}

func SuggestRepeat(source *ImageBlock, last *ImageBlock, encoder *FrameEncoder, cont bool) *EncodeSuggestion {
	return &EncodeSuggestion{
		Encoding:  ENC_REPEAT,
		MetaData:  nil,
		PixelData: nil,
		First:     !cont,
		Score:     CompareBlocks(source, last, encoder.pal),
		Result:    last,
	}
}

func SuggestSolid(source *ImageBlock, encoder *FrameEncoder) *EncodeSuggestion {
	color := -1
	score := math.MaxFloat64

	for i, palcolor := range encoder.pal {
		var scoreacc float64 = 0
		for _, pixel := range source {
			scoreacc += palcolor.ToFloatColor().Difference(encoder.pal[pixel].ToFloatColor())
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

func SuggestSolidCont(source *ImageBlock, encoder *FrameEncoder) *EncodeSuggestion {
	colorInd := encoder.GetLastSuggestion().MetaData[0]
	color := encoder.pal[colorInd]
	var score float64 = 0

	for _, pixel := range source {
		score += color.ToFloatColor().Difference(encoder.pal[pixel].ToFloatColor())
	}

	resultBlock := &ImageBlock{}
	for i := range resultBlock {
		resultBlock[i] = colorInd
	}

	result := &EncodeSuggestion{
		Encoding:  ENC_SOLID,
		MetaData:  nil,
		PixelData: nil,
		First:     false,
		Score:     score,
		Result:    resultBlock,
	}

	return result
}

func getSubColor(source int, pal Palette, subpal []int) int {
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

func applySubpal(source *ImageBlock, pal Palette, subpal []int) (data *ImageBlock, result *ImageBlock) {
	data = &ImageBlock{}
	result = &ImageBlock{}

	for i, color := range source {
		newColor := getSubColor(color, pal, subpal)
		data[i] = newColor
		result[i] = subpal[newColor]
	}
	return
}

func calcSubpal(source *ImageBlock, pal Palette, colorNum int) []int {
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

func SuggestSubColor(source *ImageBlock, encoder *FrameEncoder, encoding byte) *EncodeSuggestion {
	colorNum, _ := encodingToColors(encoding)
	data := calcSubpal(source, encoder.pal, colorNum)
	pixels, resultBlock := applySubpal(source, encoder.pal, data)
	score := CompareBlocks(source, resultBlock, encoder.pal)

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

func SuggestSubColorCont(source *ImageBlock, encoder *FrameEncoder) *EncodeSuggestion {
	var subpal []int
	encoding := encoder.GetLastSuggestion().BlockType
	if encoding == ENC_PAL2 || encoding == ENC_PAL4 || encoding == ENC_PAL8 {
		subpal = encoder.GetLastSuggestion().MetaData
	} else if encoding == ENC_PAL2_CACHE || encoding == ENC_PAL4_CACHE || encoding == ENC_PAL8_CACHE {
		_, cacheind := encodingToColors(encoding)
		subpal = encoder.palcache[cacheind].Pals[encoder.GetLastSuggestion().MetaData[0]]
	}
	pixels, resultBlock := applySubpal(source, encoder.pal, subpal)
	score := CompareBlocks(source, resultBlock, encoder.pal)

	result := &EncodeSuggestion{
		Encoding:  encoding,
		MetaData:  nil,
		PixelData: (*pixels)[:],
		First:     false,
		Score:     score,
		Result:    resultBlock,
	}

	return result
}

func SuggestSubColorCache(source *ImageBlock, encoder *FrameEncoder, encoding byte) *EncodeSuggestion {
	minscore := math.MaxFloat64
	bestIndex := 0
	_, cacheInd := encodingToColors(encoding)
	var bestPixels *ImageBlock
	var bestResult *ImageBlock
	for i, subpal := range encoder.palcache[cacheInd].GetPals() {
		pixels, resultBlock := applySubpal(source, encoder.pal, subpal)
		score := CompareBlocks(source, resultBlock, encoder.pal)
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

//region CHOOSERS

func ChooseSkip(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.lastFrame == nil {
		return nil
	}
	return SuggestSkip(input, index, encoder, false)
}

func ChooseSkipCont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_SKIP {
		return nil
	}
	return SuggestSkip(input, index, encoder, true)
}

func ChooseRepeat(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if index == 0 {
		return nil
	}
	return SuggestRepeat(input, prev, encoder, false)
}

func ChooseRepeatCont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_REPEAT {
		return nil
	}
	return SuggestRepeat(input, prev, encoder, true)
}

func ChooseSolid(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	return SuggestSolid(input, encoder)
}

func ChooseSolidCont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_SOLID {
		return nil
	}
	return SuggestSolidCont(input, encoder)
}

func ChoosePal2(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	return SuggestSubColor(input, encoder, ENC_PAL2)
}

func ChoosePal2Cont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_PAL2 {
		return nil
	}
	return SuggestSubColorCont(input, encoder)
}

func ChoosePal2Cache(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.palcache[0].Count == 0 {
		return nil
	}
	return SuggestSubColorCache(input, encoder, ENC_PAL2_CACHE)
}

func ChoosePal2CacheCont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_PAL2 {
		return nil
	}
	return SuggestSubColorCont(input, encoder)
}

func ChoosePal4(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	return SuggestSubColor(input, encoder, ENC_PAL4)
}

func ChoosePal4Cont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_PAL4 {
		return nil
	}
	return SuggestSubColorCont(input, encoder)
}

func ChoosePal4Cache(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.palcache[1].Count == 0 {
		return nil
	}
	return SuggestSubColorCache(input, encoder, ENC_PAL4_CACHE)
}

func ChoosePal4CacheCont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_PAL4 {
		return nil
	}
	return SuggestSubColorCont(input, encoder)
}

func ChoosePal8(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	return SuggestSubColor(input, encoder, ENC_PAL8)
}

func ChoosePal8Cont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_PAL8 {
		return nil
	}
	return SuggestSubColorCont(input, encoder)
}

func ChoosePal8Cache(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.palcache[2].Count == 0 {
		return nil
	}
	return SuggestSubColorCache(input, encoder, ENC_PAL8_CACHE)
}

func ChoosePal8CacheCont(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	if encoder.GetLastSuggestion() == nil || encoder.GetLastSuggestion().BlockType != ENC_PAL8 {
		return nil
	}
	return SuggestSubColorCont(input, encoder)
}

func ChooseRaw(input *ImageBlock, prev *ImageBlock, index int, encoder *FrameEncoder) *EncodeSuggestion {
	return SuggestRaw(input, encoder.GetLastSuggestion() != nil && encoder.GetLastSuggestion().BlockType == ENC_RAW)
}

//endregion
