package main

import (
	"math"
	"math/rand"
	"sort"
)

type ColorCalcMini struct {
	points    [16]ColorPoint
	centroids []FloatColor

	colors int

	totalDistance float64
	pointsChanged int

	bestError   float64
	bestPalette []FloatColor
	bestAtt     int
	errors      []float64

	maxSteps   int
	maxAttempt int
}

func NewColorCalcMini(colors int, steps int, attempts int) *ColorCalcMini {
	if colors > 8 {
		colors = 8
	}
	if colors < 1 {
		colors = 1
	}
	return &ColorCalcMini{colors: colors, maxSteps: steps, maxAttempt: attempts}
}

func (cc *ColorCalcMini) Input(block *ImageBlock, pal Palette) {
	for i, col := range block {
		cc.points[i] = ColorPoint{
			color:    pal[col].ToFloatColor(),
			segment:  0,
			count:    1,
			distance: math.MaxFloat64,
		}
	}
}

func (cc *ColorCalcMini) initCentroids() {
	centInd := 0
	swapPoints(&cc.points[0], &cc.points[rand.Uint64()%16])
	for centInd < cc.colors-1 {
		var sum float64 = 0
		for i := uint64(centInd + 1); i < 16; i++ {
			sum += cc.points[i].pointDistance(&cc.points[centInd])
		}
		rnd := rand.Float64() * sum
		centInd++
		sum = 0
		next := uint64(16 - 1)
		for i := uint64(centInd + 1); i < 16; i++ {
			sum += cc.points[i].distance
			if sum > rnd {
				next = i
				break
			}
		}
		swapPoints(&cc.points[centInd], &cc.points[next])
	}

	cc.centroids = make([]FloatColor, cc.colors)
	for i := 0; i < cc.colors; i++ {
		cc.centroids[i] = cc.points[i].color
	}
}

func (cc *ColorCalcMini) calcCentroids() {
	newCentroids := make([]FloatColor, cc.colors)
	sizes := make([]uint64, cc.colors)
	for _, point := range cc.points {
		sizes[point.segment] += point.count
		c := &newCentroids[point.segment]
		c.R += point.color.R * float64(point.count)
		c.G += point.color.G * float64(point.count)
		c.B += point.color.B * float64(point.count)
	}
	cc.totalDistance = 0
	for i := range cc.centroids {
		if sizes[i] == 0 {
			continue
		}
		size := float64(sizes[i])
		newCentroids[i].R /= size
		newCentroids[i].G /= size
		newCentroids[i].B /= size
		cc.totalDistance += math.Sqrt(newCentroids[i].Distance(cc.centroids[i]))
		cc.centroids[i] = newCentroids[i]
	}
}

func (cc *ColorCalcMini) calcSegments() {
	cc.pointsChanged = 0

	for i := range cc.points {
		oldSeg := cc.points[i].segment
		newSeg := oldSeg
		minDist := cc.points[i].color.Distance(cc.centroids[oldSeg])
		for c := range cc.centroids {
			dist := cc.points[i].color.Distance(cc.centroids[c])
			if dist < minDist {
				minDist = dist
				newSeg = c
			}
		}
		if oldSeg != newSeg {
			cc.points[i].segment = newSeg
			cc.pointsChanged++
		}
	}
}

func (cc *ColorCalcMini) CalcError() float64 {
	score := float64(0)
	for _, point := range cc.points {
		score += math.Sqrt(point.color.Distance(cc.centroids[point.segment])) * float64(point.count)
	}
	return score
}

func (cc *ColorCalcMini) Run() {
	cc.errors = make([]float64, 0, cc.maxAttempt)
	for a := 1; a < cc.maxAttempt+1; a++ {
		cc.initCentroids()
		for i := 1; i < cc.maxSteps+1; i++ {
			cc.calcSegments()
			if cc.pointsChanged == 0 {
				break
			}
			cc.calcCentroids()
		}
		cc.calcSegments()
		colorErr := cc.CalcError()
		if a == 1 || colorErr < cc.bestError {
			cc.bestAtt = a
			cc.bestError = colorErr
			cc.bestPalette = cc.calcPalette()
		}
		cc.errors = append(cc.errors, colorErr)
	}
}

func (cc *ColorCalcMini) calcPalette() []FloatColor {
	result := make([]FloatColor, cc.colors)
	copy(result, cc.centroids)
	return result
}

func (cc *ColorCalcMini) GetSubPal(pal Palette) []int {
	result := make([]int, cc.colors)
	for i, col := range cc.bestPalette {
		result[i] = pal.GetFloatColorIndex(col)
	}
	sort.Ints(result)
	return result
}
