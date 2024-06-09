package main

import "math"

type PalComp struct {
	pal        Palette
	floatpal   []FloatColor
	diffmatrix [][]float64
	lumas      []float64
}

func NewPalComp(pal Palette) *PalComp {
	floatpal := make([]FloatColor, pal.Len())
	lumas := make([]float64, pal.Len())
	diffmat := make([][]float64, pal.Len())
	for i, c := range pal {
		fc := c.ToFloatColor()
		floatpal[i] = fc
		lumas[i] = fc.Luma()
		diffmat[i] = make([]float64, pal.Len())
		for j, c2 := range pal {
			diffmat[i][j] = fc.Difference(c2.ToFloatColor())
		}
	}
	return &PalComp{
		pal:        pal,
		floatpal:   floatpal,
		lumas:      lumas,
		diffmatrix: diffmat,
	}
}

func (pc *PalComp) CompareColors(c1 int, c2 int) float64 {
	return pc.diffmatrix[c1][c2]
}

func (pc *PalComp) colorDiff(color FloatColor, luma float64, id int) float64 {
	other := pc.floatpal[id]
	diffR := color.R - other.R
	diffG := color.G - other.G
	diffB := color.B - other.B
	diffcolor := diffR*diffR*0.299 + diffG*diffG*0.587 + diffB*diffB*0.114
	diffluma := luma - pc.lumas[id]
	return diffcolor*0.75 + diffluma*diffluma
}

func (pc *PalComp) GetColorIndex(color FloatColor) int {
	luma := color.Luma()
	var minDist float64 = math.MaxFloat64
	minIndex := 0
	for i := range pc.floatpal {
		dist := pc.colorDiff(color, luma, i)
		//dist := color.Difference(pc.floatpal[i])
		if dist < minDist {
			minDist = dist
			minIndex = i
		}
	}
	return minIndex
}

func (pc *PalComp) ToFloatColor(index int) FloatColor {
	return pc.floatpal[index]
}
