package main

import (
	"runtime"
	"sort"
	"sync"
)

type DitheringMethod interface {
	Init(pal Palette, pc *PalComp, width int, height int)
	Process(imageData []IntColor, pal Palette) []int
}

//region POSTERIZE

type PosterizeDithering struct{}

func (dither *PosterizeDithering) Init(pal Palette, pc *PalComp, width int, height int) {}
func (dither *PosterizeDithering) Process(imageData []IntColor, pal Palette) []int {
	idata := make([]int, len(imageData))
	for i := range idata {
		idata[i] = pal.GetFloatColorIndex(imageData[i].ToFloatColor())
	}
	return idata
}

//endregion

//region FLOYD–STEINBERG

func addError(dst *FloatColor, err float64) {
	dst.R = clipFloat(dst.R + err)
	dst.G = clipFloat(dst.G + err)
	dst.B = clipFloat(dst.B + err)
}

type FSDithering struct {
	fdata  []FloatColor
	width  int
	height int
}

func (dither *FSDithering) Init(pal Palette, pc *PalComp, width int, height int) {
	dither.fdata = make([]FloatColor, width*height)
	dither.width = width
	dither.height = height
}

func (dither *FSDithering) Process(imageData []IntColor, pal Palette) []int {
	idata := make([]int, dither.width*dither.height)

	for i := range dither.fdata {
		dither.fdata[i] = imageData[i].ToFloatColor()
	}

	for y := 0; y < dither.height; y++ {
		for x := 0; x < dither.width; x++ {
			index := y*dither.width + x
			oldColor := dither.fdata[index]
			newColorIndex := pal.GetFloatColorIndex(oldColor)
			newColor := pal[newColorIndex].ToFloatColor()
			idata[index] = newColorIndex
			dither.fdata[index] = newColor
			colError := (oldColor.R - newColor.R + oldColor.G - newColor.G + oldColor.B - newColor.B) / 3
			if x < dither.width-1 {
				addError(&dither.fdata[y*dither.width+x+1], colError*7.0/16.0)
			}
			if y < dither.height-1 {
				if x > 0 {
					addError(&dither.fdata[(y+1)*dither.width+x-1], colError*3.0/16.0)
				}
				addError(&dither.fdata[(y+1)*dither.width+x], colError*5.0/16.0)
				if x < dither.width-1 {
					addError(&dither.fdata[(y+1)*dither.width+x+1], colError*1.0/16.0)
				}
			}
		}
	}
	return idata
}

//endregion

//region PATTERN

type PatternDithering struct {
	width     int
	height    int
	pattern   *Pattern
	fdata     []FloatColor
	workers   int
	rangeSize int
	treshold  float64
	pc        *PalComp
}

func NewPatternDithering(pattern *Pattern, workers int, treshold float64) *PatternDithering {
	if workers < 0 {
		workers = runtime.NumCPU()
	}
	return &PatternDithering{
		pattern:  pattern,
		workers:  workers,
		treshold: treshold,
	}
}

func (dither *PatternDithering) Init(pal Palette, pc *PalComp, width int, height int) {
	dither.width = width
	dither.height = height
	dither.pattern = dither.pattern.Reshape(width, height)
	dither.fdata = make([]FloatColor, width*height)
	dither.rangeSize = width * height / dither.workers
	dither.pc = pc
}

func (dither *PatternDithering) Process(imageData []IntColor, pal Palette) []int {
	idata := make([]int, dither.width*dither.height)

	for i := range dither.fdata {
		dither.fdata[i] = imageData[i].ToFloatColor()
	}

	var wg sync.WaitGroup

	workerFunc := func(wdata []FloatColor, widata []int, wpattern []int) {
		candidates := make([]int, dither.pattern.Order)
		for p := range wdata {
			cerr := FloatColor{0, 0, 0}
			for i := range candidates {
				attempt := wdata[p]
				attempt.R = clipFloat(attempt.R + cerr.R*dither.treshold)
				attempt.G = clipFloat(attempt.G + cerr.G*dither.treshold)
				attempt.B = clipFloat(attempt.B + cerr.B*dither.treshold)
				//colorIndex := pal.GetFloatColorIndex(attempt)
				colorIndex := dither.pc.GetColorIndex(attempt)
				candidates[i] = colorIndex
				candidate := dither.pc.ToFloatColor(colorIndex) //pal[colorIndex].ToFloatColor()
				cerr.R += wdata[p].R - candidate.R
				cerr.G += wdata[p].G - candidate.G
				cerr.B += wdata[p].B - candidate.B
			}
			sort.Ints(candidates[:])
			widata[p] = candidates[wpattern[p]]
		}
		wg.Done()
	}

	for i := 0; i < dither.workers-1; i++ {
		rangeStart := i * dither.rangeSize
		rangeEnd := (i + 1) * dither.rangeSize
		wg.Add(1)
		go workerFunc(dither.fdata[rangeStart:rangeEnd], idata[rangeStart:rangeEnd], dither.pattern.Data[rangeStart:rangeEnd])
	}
	rangeStart := (dither.workers - 1) * dither.rangeSize
	wg.Add(1)
	go workerFunc(dither.fdata[rangeStart:], idata[rangeStart:], dither.pattern.Data[rangeStart:])

	wg.Wait()
	return idata
}

//endregion

func FindDithering(name string) DitheringMethod {
	switch name {
	case "none":
		return &PosterizeDithering{}
	case "fs":
		return &FSDithering{}
	case "pat8":
		return NewPatternDithering(bayerMatrixBig, -1, 0.5)
	case "pat4":
		return NewPatternDithering(bayerMatrixSmall, -1, 0.5)
	default:
		return nil
	}
}
