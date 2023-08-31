package main

import (
	"runtime"
	"sort"
	"sync"
)

var ditherMap = [8][8]int{
	{0, 32, 8, 40, 2, 34, 10, 42},
	{48, 16, 56, 24, 50, 18, 58, 26},
	{12, 44, 4, 36, 14, 46, 6, 38},
	{60, 28, 52, 20, 62, 30, 54, 22},
	{3, 35, 11, 43, 1, 33, 9, 41},
	{51, 19, 59, 27, 49, 17, 57, 25},
	{15, 47, 7, 39, 13, 45, 5, 37},
	{63, 31, 55, 23, 61, 29, 53, 21}}

var ditherMapSmall = [4][4]int{
	{0, 8, 2, 10},
	{12, 4, 14, 6},
	{3, 11, 1, 9},
	{15, 7, 13, 5}}

type ImageDithering func(imageData []IntColor, pal Palette, width, height int) []int

func DitheringPosterize(imageData []IntColor, pal Palette, width, height int) []int {
	idata := make([]int, len(imageData))
	for i := range idata {
		//idata[i] = pal.GetIntColorIndex(imageData[i])
		idata[i] = pal.GetFloatColorIndex(imageData[i].ToFloatColor())
	}
	return idata
}

func addError(dst *FloatColor, err float64) {
	dst.R = clipFloat(dst.R + err)
	dst.G = clipFloat(dst.G + err)
	dst.B = clipFloat(dst.B + err)
}

func DitheringFS(imageData []IntColor, pal Palette, width, height int) []int {
	data := make([]FloatColor, width*height)
	idata := make([]int, width*height)

	for i := range data {
		data[i] = imageData[i].ToFloatColor()
	}

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*width + x
			oldColor := data[index]
			newColorIndex := pal.GetFloatColorIndex(oldColor)
			newColor := pal[newColorIndex].ToFloatColor()
			idata[index] = newColorIndex
			data[index] = newColor
			colError := (oldColor.R - newColor.R + oldColor.G - newColor.G + oldColor.B - newColor.B) / 3
			if x < width-1 {
				addError(&data[y*width+x+1], colError*7.0/16.0)
			}
			if y < height-1 {
				if x > 0 {
					addError(&data[(y+1)*width+x-1], colError*3.0/16.0)
				}
				addError(&data[(y+1)*width+x], colError*5.0/16.0)
				if x < width-1 {
					addError(&data[(y+1)*width+x+1], colError*1.0/16.0)
				}
			}
		}
	}
	return idata
}

func DitheringPattern8(imageData []IntColor, pal Palette, width, height int) []int {
	data := make([]FloatColor, width*height)
	idata := make([]int, width*height)
	pattern := make([]int, width*height)

	for i := range data {
		data[i] = imageData[i].ToFloatColor()
	}

	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pattern[index] = ditherMap[x%8][y%8]
			index++
		}
	}

	var treshold float64 = 0.5
	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	rangeSize := len(data) / workers

	workerFunc := func(wdata []FloatColor, widata []int, wpattern []int) {
		var candidates [8 * 8]int
		for p := range wdata {
			cerr := FloatColor{0, 0, 0}
			for i := range candidates {
				attempt := wdata[p]
				attempt.R = clipFloat(attempt.R + cerr.R*treshold)
				attempt.G = clipFloat(attempt.G + cerr.G*treshold)
				attempt.B = clipFloat(attempt.B + cerr.B*treshold)
				colorIndex := pal.GetFloatColorIndex(attempt)
				candidates[i] = colorIndex
				candidate := pal[colorIndex].ToFloatColor()
				cerr.R += wdata[p].R - candidate.R
				cerr.G += wdata[p].G - candidate.G
				cerr.B += wdata[p].B - candidate.B
			}
			sort.Ints(candidates[:])
			widata[p] = candidates[wpattern[p]]
		}
		wg.Done()
	}

	for i := 0; i < workers-1; i++ {
		rangeStart := i * rangeSize
		rangeEnd := (i + 1) * rangeSize
		wg.Add(1)
		go workerFunc(data[rangeStart:rangeEnd], idata[rangeStart:rangeEnd], pattern[rangeStart:rangeEnd])
	}
	rangeStart := (workers - 1) * rangeSize
	wg.Add(1)
	go workerFunc(data[rangeStart:], idata[rangeStart:], pattern[rangeStart:])

	wg.Wait()
	return idata
}

func DitheringPattern4(imageData []IntColor, pal Palette, width, height int) []int {
	data := make([]FloatColor, width*height)
	idata := make([]int, width*height)
	pattern := make([]int, width*height)

	for i := range data {
		data[i] = imageData[i].ToFloatColor()
	}

	index := 0
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pattern[index] = ditherMapSmall[x%4][y%4]
			index++
		}
	}

	var treshold float64 = 0.5
	var wg sync.WaitGroup
	workers := runtime.NumCPU()
	rangeSize := len(data) / workers

	workerFunc := func(wdata []FloatColor, widata []int, wpattern []int) {
		var candidates [4 * 4]int
		for p := range wdata {
			cerr := FloatColor{0, 0, 0}
			for i := range candidates {
				attempt := wdata[p]
				attempt.R = clipFloat(attempt.R + cerr.R*treshold)
				attempt.G = clipFloat(attempt.G + cerr.G*treshold)
				attempt.B = clipFloat(attempt.B + cerr.B*treshold)
				colorIndex := pal.GetFloatColorIndex(attempt)
				candidates[i] = colorIndex
				candidate := pal[colorIndex].ToFloatColor()
				cerr.R += wdata[p].R - candidate.R
				cerr.G += wdata[p].G - candidate.G
				cerr.B += wdata[p].B - candidate.B
			}
			sort.Ints(candidates[:])
			widata[p] = candidates[wpattern[p]]
		}
		wg.Done()
	}

	for i := 0; i < workers-1; i++ {
		rangeStart := i * rangeSize
		rangeEnd := (i + 1) * rangeSize
		wg.Add(1)
		go workerFunc(data[rangeStart:rangeEnd], idata[rangeStart:rangeEnd], pattern[rangeStart:rangeEnd])
	}
	rangeStart := (workers - 1) * rangeSize
	wg.Add(1)
	go workerFunc(data[rangeStart:], idata[rangeStart:], pattern[rangeStart:])

	wg.Wait()
	return idata
}

func FindDithering(name string) ImageDithering {
	switch name {
	case "none":
		return DitheringPosterize
	case "fs":
		return DitheringFS
	case "pat8":
		return DitheringPattern8
	case "pat4":
		return DitheringPattern4
	default:
		return nil
	}
}
