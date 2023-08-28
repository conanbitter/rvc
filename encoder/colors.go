package main

import "math"

type FloatColor struct {
	R, G, B float64
}

type IntColor struct {
	R, G, B int
}

type Palette []IntColor

func clipFloat(val float64) float64 {
	if val > 1.0 {
		return 1.0
	}
	if val < 0.0 {
		return 0.0
	}
	return val
}

func clipInt(val int) int {
	if val > 255 {
		return 255
	}
	if val < 0 {
		return 0
	}
	return val
}

//===== FLOAT COLOR =======

func (color FloatColor) Normalized() FloatColor {
	return FloatColor{
		clipFloat(color.R),
		clipFloat(color.G),
		clipFloat(color.B)}
}

func (color FloatColor) Distance(other FloatColor) float64 {
	return math.Sqrt((color.R-other.R)*(color.R-other.R) +
		(color.G-other.G)*(color.G-other.G) +
		(color.B-other.B)*(color.B-other.B))
}

func (color FloatColor) DistanceSquared(other FloatColor) float64 {
	return (color.R-other.R)*(color.R-other.R) +
		(color.G-other.G)*(color.G-other.G) +
		(color.B-other.B)*(color.B-other.B)
}

func (color FloatColor) ToIntColor() IntColor {
	norm := color.Normalized()
	return IntColor{
		clipInt(int(math.Round(norm.R * 255))),
		clipInt(int(math.Round(norm.G * 255))),
		clipInt(int(math.Round(norm.B * 255)))}
}

//===== INT COLOR =======

func (color IntColor) Normalized() IntColor {
	return IntColor{
		clipInt(color.R),
		clipInt(color.G),
		clipInt(color.B)}
}

func (color IntColor) Luma() float64 {
	return 0.2126*float64(color.R) + 0.7152*float64(color.G) + 0.0722*float64(color.B)
}

func (color IntColor) ToFloatColor() FloatColor {
	return FloatColor{float64(color.R) / 255.0, float64(color.G) / 255.0, float64(color.B) / 255.0}
}
